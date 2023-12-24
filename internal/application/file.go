package application

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/ywak/zouni/internal"
)

// あるパッケージに含まれるソースコードのファイルをあらわします。
type File struct {
	// このソースコードが含まれるパッケージ
	pkg *Package

	// このソースコードのファイル名
	fileName string

	tokenFile *token.File

	astFile *ast.File

	// このファイルで定義されたselectorと実際のimport
	selector2path map[string]string
}

func NewFile(pkg *Package, fileName string) *File {
	file := File{
		pkg:      pkg,
		fileName: fileName,
	}

	return &file
}

func (file *File) Parse(fset *token.FileSet) error {
	bpkg := file.pkg.bpkg
	var fname string
	if bpkg.Name == "main" {
		fname = file.fileName
	} else {
		fname = filepath.Join(bpkg.ImportPath, file.fileName)
	}

	path := filepath.Join(bpkg.Dir, file.fileName)
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	astFile, err := parser.ParseFile(fset, fname, buf, parser.ParseComments)
	if err != nil {
		return err
	}

	file.astFile = astFile
	file.tokenFile = fset.File(astFile.Package)
	file.selector2path = map[string]string{}
	for _, iSpec := range astFile.Imports {
		internal.Log.Printf("import spec, Name = '%v', Path = '%v'", iSpec.Name, iSpec.Path)
		path := iSpec.Path.Value
		var name string
		if iSpec.Name == nil {
			path = TrimQuote(path)
			name = path[(strings.LastIndexByte(path, '/') + 1):]
		} else {
			name = iSpec.Name.Name
		}

		file.selector2path[name] = TrimQuote(iSpec.Path.Value)
	}
	return nil
}

func (file *File) Name() string {
	return file.astFile.Name.Name
}

func (file *File) Decls() []ast.Decl {
	return file.astFile.Decls
}

func (file *File) FindPackageBySelector(selector string) string {
	if path, ex := file.selector2path[selector]; ex {
		return path
	}

	internal.Log.Printf("path is not found for '%v' in %v", selector, file.selector2path)
	return ""
}
