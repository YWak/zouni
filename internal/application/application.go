package application

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"path/filepath"

	"github.com/ywak/zouni/internal/config"
)

// importを追いかけて必要なパッケージをすべて読み込む
// パッケージに存在する宣言をすべて読み込む
// main関数から構文解析木をたどって、実際に使われる宣言を取得する。使われるとは
// 1. main関数で使用されている型、変数、関数および、それらを再帰的に辿って発見した型
// 2. 使用されている構造体型に定義されたメソッド

type Application struct {
	fset *token.FileSet

	// mainを含むパッケージ
	root *build.Package

	// パッケージ名とパッケージの対応
	pkgs map[string]*Package

	// tokenのキャッシュ
	tokens map[ast.Decl]bool

	config *config.ZouniConfig
}

// ルートから関連する全パッケージを読み出します。
func Load(cfg *config.ZouniConfig, rootDir string) (*Application, error) {
	build.Default.Dir = rootDir // TODO これを変更する方法はある？

	fset := token.NewFileSet()
	app := Application{
		fset:   fset,
		pkgs:   map[string]*Package{},
		tokens: map[ast.Decl]bool{},
		config: cfg,
	}
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	cfg.Logger().Printf("rootDir = %s", rootDir)

	rootPkg, err := build.ImportDir(rootDir, 0)
	if err != nil {
		return nil, fmt.Errorf("import failed on %s. %v", rootDir, err)
	}

	cfg.Logger().Printf("root pkg = %v", rootPkg)
	app.root = rootPkg
	app.pkgs["main"], err = NewPackage(cfg, fset, rootPkg)
	if err != nil {
		return nil, err
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
				return nil, fmt.Errorf("failure on getting import path to %s %s. %v", importValue, bpkg.Dir, err)
			}

			nextPkg, err := build.ImportDir(importDir, 0)
			if err != nil {
				return nil, fmt.Errorf("import failure for %s. %v", importValue, err)
			}

			cfg.Logger().Printf("next package = %v", nextPkg)
			pkg, err := NewPackage(cfg, fset, nextPkg)
			if err != nil {
				return nil, err
			}
			app.pkgs[importValue] = pkg

			pkgQueue = append(pkgQueue, nextPkg)
		}

		cfg.Logger().Printf("%v", *bpkg)
	}

	return &app, nil
}

func (app *Application) getImportPath(path string, dir string) (string, error) {
	bpkg, err := build.Import(path, dir, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("error on app.getImportPath(%s, %s). %v", path, dir, err)
	}
	return bpkg.Dir, nil
}

func (app *Application) String() string {
	buf := bytes.NewBuffer(make([]byte, 1024*1024))
	fmt.Fprintf(buf, "Application {\n")

	fmt.Fprintf(buf, "  pkgs: {\n")
	for name, pkg := range app.pkgs {
		fmt.Fprintf(buf, "    '%s':\n%s,\n", name, pkg.IndentedString(6))
	}
	fmt.Fprintf(buf, "  }\n")

	fmt.Fprintf(buf, "}\n")

	return buf.String()
}
