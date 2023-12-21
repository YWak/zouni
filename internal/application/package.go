package application

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/ywak/zouni/internal/config"
)

type Package struct {
	cfg   *config.ZouniConfig
	bpkg  *build.Package
	files map[string]*ast.File
}

func NewPackage(cfg *config.ZouniConfig, fset *token.FileSet, bpkg *build.Package) (*Package, error) {
	pkg := Package{
		cfg:   cfg,
		bpkg:  bpkg,
		files: map[string]*ast.File{},
	}

	files := []string{}
	files = append(files, bpkg.GoFiles...)
	files = append(files, bpkg.CgoFiles...)

	for _, f := range files {
		var fname string
		if bpkg.Name == "main" {
			fname = f
		} else {
			fname = filepath.Join(bpkg.ImportPath, f)
		}

		path := filepath.Join(bpkg.Dir, f)
		if _, ex := pkg.files[path]; !ex {
			buf, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			astFile, err := parser.ParseFile(fset, fname, buf, parser.ParseComments)
			if err != nil {
				return nil, err
			}
			pkg.files[path] = astFile
			cfg.Logger().Printf("parsed %s", path)
		}
	}

	return &pkg, nil
}

func (pkg *Package) Decls() []ast.Decl {
	decls := []ast.Decl{}
	for _, file := range pkg.files {
		decls = append(decls, file.Decls...)
	}

	return decls
}

func (pkg *Package) String() string {
	return pkg.IndentedString(0)
}

func (pkg *Package) IndentedString(indentSize int) string {
	indent := strings.Repeat(" ", indentSize)

	buf := bytes.NewBuffer(make([]byte, 1024*1024))
	fmt.Fprintf(buf, "%sPackage {\n", indent)
	fmt.Fprintf(buf, "%s  bpkg: '%s',\n", indent, pkg.bpkg.Name)
	fmt.Fprintf(buf, "%s  files: {\n", indent)
	for name, file := range pkg.files {
		fmt.Fprintf(buf, "%s    '%s': '%s',\n", indent, name, file.Name)
	}
	fmt.Fprintf(buf, "%s  }\n", indent)
	fmt.Fprintf(buf, "%s}", indent)

	return buf.String()
}

var _ fmt.Stringer = &Package{}
