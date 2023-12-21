package application

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"path/filepath"

	"github.com/ywak/zouni/internal"
	"github.com/ywak/zouni/internal/config"
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

	config *config.ZouniConfig
}

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

func Load(cfg *config.ZouniConfig, rootDir string) (*Application, error) {
	build.Default.Dir = rootDir // TODO これを変更する方法はある？

	fset := token.NewFileSet()
	app := Application{
		fset:   fset,
		config: cfg,
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

	app.config.Logger().Printf("rootDir = %s", rootDir)

	rootPkg, err := build.ImportDir(rootDir, 0)
	if err != nil {
		return fmt.Errorf("import failed on %s. %v", rootDir, err)
	}

	app.config.Logger().Printf("root pkg = %v", rootPkg)
	app.root = rootPkg
	app.pkgs = map[string]*Package{}
	app.pkgs["main"], err = NewPackage(app.config, app.fset, rootPkg)
	if err != nil {
		return err
	}

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

			app.config.Logger().Printf("next package = %v", nextPkg)
			pkg, err := NewPackage(app.config, app.fset, nextPkg)
			if err != nil {
				return err
			}
			app.pkgs[importValue] = pkg

			pkgQueue = append(pkgQueue, nextPkg)
		}

		app.config.Logger().Printf("%v", *bpkg)
	}

	return nil
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
