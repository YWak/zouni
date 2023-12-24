package application

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"path/filepath"

	"github.com/ywak/zouni/internal"
)

// - importを追いかけて必要なパッケージをすべて読み込む
// - パッケージに存在する宣言をすべて読み込む
// - main関数から構文解析木をたどって実際に使われる宣言を取得し、帰りがけ順で保存する。使われるとは
//  1. main関数で使用されている型、変数、関数および、それらを再帰的に辿って発見した型
//  2. 使用されている構造体型に定義されたメソッド
type Application struct {
	fset *token.FileSet

	// mainを含むパッケージ
	root *build.Package

	// パッケージ名とパッケージの対応
	pkgs map[string]*Package

	decls []ast.Decl

	// token.Fileがあらわすファイルと同一のファイルに対応するast.File
	f2f map[*token.File]*ast.File
}

// 指定されたアプリケーションの全構造を読み込みます。
func Load(fset *token.FileSet, rootDir string) (*Application, error) {
	build.Default.Dir = rootDir // TODO これを変更する方法はある？

	app := Application{
		fset: fset,
	}
	err := app.loadPackages(rootDir)
	if err != nil {
		return nil, err
	}

	err = app.loadDecls()
	if err != nil {
		return nil, err
	}

	return &app, nil
}

// ルートから関連する全パッケージを読み出します。
func (app *Application) loadPackages(rootDir string) error {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return err
	}

	app.logf("rootDir = %s", rootDir)

	rootPkg, err := build.ImportDir(rootDir, 0)
	if err != nil {
		return fmt.Errorf("import failed on %s. %v", rootDir, err)
	}

	app.pkgs = map[string]*Package{}
	app.f2f = map[*token.File]*ast.File{}

	_, err = app.addPackageFrom("main", rootPkg)
	if err != nil {
		return err
	}

	app.logf("root pkg = %v", rootPkg)
	app.root = rootPkg

	// ファイルをすべてたどる
	pkgQueue := []*build.Package{rootPkg}
	for len(pkgQueue) > 0 {
		bpkg := pkgQueue[0]
		pkgQueue = pkgQueue[1:]

		// 全importを探す
		for _, importValue := range bpkg.Imports {
			importDir, err := app.getImportPath(importValue, bpkg.Dir)
			if err != nil {
				return fmt.Errorf("failure on getting import path to %s %s. %v", importValue, bpkg.Dir, err)
			}

			nextPkg, err := build.ImportDir(importDir, 0)
			if err != nil {
				return fmt.Errorf("import failure for %s. %v", importValue, err)
			}

			app.logf("next package = %v", nextPkg)
			_, err = app.addPackageFrom(importValue, nextPkg)
			if err != nil {
				return err
			}

			pkgQueue = append(pkgQueue, nextPkg)
		}

		app.logf("%v", *bpkg)
	}

	return nil
}

func (app *Application) addPackageFrom(name string, bpkg *build.Package) (*Package, error) {
	pkg, err := CreatePackage(app.fset, bpkg)
	if err != nil {
		return nil, err
	}
	app.pkgs[name] = pkg
	for tf, af := range pkg.f2f {
		app.f2f[tf] = af
	}

	return pkg, nil
}

func (app *Application) getImportPath(path string, dir string) (string, error) {
	bpkg, err := build.Import(path, dir, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("error on app.getImportPath(%s, %s). %v", path, dir, err)
	}
	return bpkg.Dir, nil
}

// 各パッケージから宣言を読み出します。
func (app *Application) loadDecls() error {
	for _, pkg := range app.pkgs {
		app.decls = append(app.decls, pkg.Decls()...)
	}
	return nil
}

// init関数およびmain関数を起点として、必要な宣言を登場順に返します。
func (app *Application) Dig() ([]ast.Decl, error) {
	// 宣言の登場順
	decls := map[ast.Decl]int{}
	push := func(decl ast.Decl) bool {
		if _, ex := decls[decl]; ex {
			return false
		}
		decls[decl] = len(decls)
		return true
	}

	// 起点になる関数
	roots := []*ast.FuncDecl{}

	// mainを探す
	for _, decl := range app.pkgs["main"].Decls() {
		if fun, ok := decl.(*ast.FuncDecl); ok && fun.Name.Name == "main" {
			roots = append(roots, fun)
			push(fun)
			break
		}
	}

	// 宣言をもとにして登場する型、関数、宣言を拾っていく
	for len(roots) > 0 {
		root := roots[0]
		push(root)

		roots = roots[1:]

		// rootから全nodeを辿って宣言を拾っていく。
		ast.Inspect(root, func(n ast.Node) bool {
			// app.logf("node = %v", internal.DescribeNode(app.fset, n))
			switch node := n.(type) {
			case *ast.CallExpr: // 関数呼び出しなら、呼び出される関数を必要とするので定義を保存しておく
				decl := app.findFuncDecl(node)
				if decl != nil && push(decl) {
					roots = append(roots, decl)
				}

				// 引数の解析を継続する必要がある
				return true
			default:
				return true
			}
		})
	}

	ret := make([]ast.Decl, len(decls))
	for decl, i := range decls {
		ret[i] = decl
	}

	return ret, nil
}

// 関数呼び出しで呼び出されている関数の定義を返します。
func (app *Application) findFuncDecl(node *ast.CallExpr) *ast.FuncDecl {
	// app.logf("type = %v", internal.DescribeNode(app.fset, node))
	switch fun := node.Fun.(type) {
	case *ast.Ident:
		// 呼ばなくていい？
		return nil
	case *ast.SelectorExpr:
		functionName := fun.Sel.Name
		var importName string
		switch x := fun.X.(type) {
		case *ast.Ident:
			importName = TrimQuote(x.Name)
		default:
			importName = ""
		}

		// ファイル -> selector -> import -> packageと辿る
		file := app.findFile(node)
		packageName := file.FindPackageBySelector(importName)
		pkg := app.findPackageByName(packageName)
		return pkg.FindFunctionByName(functionName)
	default:
		// app.logf("type = %v", internal.DescribeNode(app.fset, fun))
		return nil
	}
}

// nodeが格納されているファイルを返します。
func (app *Application) findFile(node ast.Node) *File {
	tf := app.fset.File(node.Pos())
	for _, pkg := range app.pkgs {
		for _, file := range pkg.files {
			if file.tokenFile == tf {
				return file
			}
		}
	}
	return nil
}

func (app *Application) findPackageByName(name string) *Package {
	if pkg, ex := app.pkgs[name]; ex {
		return pkg
	}

	internal.Log.Printf("pkg is nil for %v in %v", name, app.pkgs)
	return nil
}

// appを文字列で表現します。
func (app *Application) String() string {
	buf := bytes.NewBuffer(make([]byte, 1024*1024))
	fmt.Fprintf(buf, "Application {\n")

	fmt.Fprintf(buf, "  pkgs: {\n")
	for name, pkg := range app.pkgs {
		fmt.Fprintf(buf, "    '%s':\n%s,\n", name, pkg.IndentedString(6))
	}
	fmt.Fprintf(buf, "  }\n")

	fmt.Fprintf(buf, "  delcs: [\n")
	for _, decl := range app.decls {
		fmt.Fprintf(buf, "    %s\n", internal.DescribeDecl(app.fset, decl))
	}

	fmt.Fprintf(buf, "}\n")

	return buf.String()
}

// ログを出力するためのコンビニ関数
func (app *Application) logf(fmt string, v ...any) {
	internal.Log.Printf(fmt, v...)
}
