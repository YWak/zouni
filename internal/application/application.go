package application

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"path/filepath"

	"github.com/ywak/zouni/internal/config"
)

type Application struct {
	fset   *token.FileSet
	root   *build.Package
	pkgs   map[string]*build.Package
	tokens map[ast.Decl]bool
	config *config.ZouniConfig
}

func Load(cfg *config.ZouniConfig, rootDir string) (*Application, error) {
	build.Default.Dir = rootDir // TODO これを変更する方法はある？

	fset := token.NewFileSet()
	app := Application{
		fset:   fset,
		pkgs:   map[string]*build.Package{},
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
	app.pkgs["main"] = rootPkg

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
			app.pkgs[importValue] = nextPkg
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
