package application

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"strings"

	"github.com/ywak/zouni/internal"
)

type Package struct {
	// このパッケージに対応するパッケージ
	bpkg *build.Package

	// このパッケージに含まれるファイルのパスとファイルをあらわす構造体の対応
	files []*File

	// このパッケージに含まれるtoken.Fileとast.Fileの対応
	f2f map[*token.File]*ast.File
}

// パッケージを作成します。
func CreatePackage(fset *token.FileSet, bpkg *build.Package) (*Package, error) {
	pkg := Package{
		bpkg: bpkg,
		f2f:  map[*token.File]*ast.File{},
	}

	files := []string{}
	files = append(files, bpkg.GoFiles...)
	files = append(files, bpkg.CgoFiles...)

	for _, f := range files {
		file := NewFile(&pkg, f)
		if err := file.Parse(fset); err != nil {
			return nil, err
		}
		pkg.files = append(pkg.files, file)
	}

	return &pkg, nil
}

func (pkg *Package) Decls() []ast.Decl {
	decls := []ast.Decl{}
	for _, file := range pkg.files {
		decls = append(decls, file.Decls()...)
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
	fmt.Fprintf(buf, "%s  files: [\n", indent)
	for _, file := range pkg.files {
		fmt.Fprintf(buf, "%s    '%s',\n", indent, file.Name())
	}
	fmt.Fprintf(buf, "%s  }\n", indent)
	fmt.Fprintf(buf, "%s}", indent)

	return buf.String()
}

func (pkg *Package) FindFunctionByName(functionName string) *ast.FuncDecl {
	for _, decl := range pkg.Decls() {
		if fun, ok := decl.(*ast.FuncDecl); ok && fun.Name.Name == functionName {
			internal.Log.Printf("decl %v is found for '%s'", fun, functionName)
			return fun
		}
	}

	internal.Log.Printf("package is not found for '%s' in %v", functionName, pkg.Decls())
	return nil
}

var _ fmt.Stringer = &Package{}
