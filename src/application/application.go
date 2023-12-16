package application

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"path/filepath"

	"github.com/ywak/zouni/src/config"
)

type Application struct {
	fset *token.FileSet
	// root   *build.Package
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

	cfg.Logger().Printf("default = %v", build.Default)
	cfg.Logger().Printf("rootDir = %s", rootDir)
	rootPkg, err := build.ImportDir(rootDir, 0)
	if err != nil {
		return nil, fmt.Errorf("import failed on %s. %v", rootDir, err)
	}

	cfg.Logger().Printf("root pkg = %v", rootPkg)

	pkgQueue := []*build.Package{rootPkg}
	tokenQueue := []ast.Decl{}

	// ファイルをすべてたどる
	for len(pkgQueue) > 0 {
		bpkg := pkgQueue[0]
		pkgQueue = pkgQueue[1:]

		// 全importを探す
		for _, importValue := range bpkg.Imports {
			cfg.Logger().Println("importValue =", importValue)
			importDir, err := app.getImportPath(importValue, bpkg.Dir)
			if err != nil {
				return nil, fmt.Errorf("failure on getting import path to %s %s. %v", importValue, bpkg.Dir, err)
			}
			nextPkg, err := build.ImportDir(importDir, 0)
			if err != nil {
				return nil, fmt.Errorf("import failure for %s. %v", importValue, err)
			}
			// nextFilePath, err := app.getImportPath(importValue, bpkg.Dir)

			cfg.Logger().Printf("%v", nextPkg)
		}

		cfg.Logger().Printf("%v", *bpkg)
	}

	// 宣言をたどって全nodeを取得する
	for len(tokenQueue) > 0 {

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
