package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"testing"
)

type TestLogger struct {
	T *testing.T
}

func (l *TestLogger) Printf(fmt string, v ...any) {
	l.T.Logf(fmt, v...)
}

func (l *TestLogger) Println(v ...any) {
	l.T.Log(v...)
}

func DescribeDecl(fset *token.FileSet, decl ast.Decl) string {
	switch t := decl.(type) {
	case *ast.GenDecl:
		switch t.Tok {
		case token.IMPORT:
			return fmt.Sprintf("import %s(%s)", describeSpec(fset, t.Specs[0]), fset.PositionFor(t.Pos(), true))
		case token.CONST:
			return fmt.Sprintf("const %s(%s)", describeSpec(fset, t.Specs[0]), fset.PositionFor(t.Pos(), true))
		case token.TYPE:
			return fmt.Sprintf("type %s(%s)", describeSpec(fset, t.Specs[0]), fset.PositionFor(t.Pos(), true))
		case token.VAR:
			return fmt.Sprintf("var %s(%s)", describeSpec(fset, t.Specs[0]), fset.PositionFor(t.Pos(), true))
		}
	case *ast.FuncDecl:
		return fmt.Sprintf("func %s(%s)", t.Name.Name, fset.PositionFor(t.Pos(), true))
	}

	buf := bytes.NewBuffer(make([]byte, 1024))
	ast.Fprint(buf, fset, decl, nil)
	return buf.String()
}

func describeSpec(fset *token.FileSet, spec ast.Spec) string {
	switch sp := spec.(type) {
	case *ast.ImportSpec:
		return sp.Path.Value
	case *ast.TypeSpec:
		switch sp.Type.(type) {
		case *ast.InterfaceType:
			return fmt.Sprintf("%s interface", sp.Name.Name)
		default:
			return fmt.Sprintf("%s %s", sp.Name.Name, sp.Type)
		}
	}

	buf := bytes.NewBuffer(make([]byte, 1024))
	ast.Fprint(buf, fset, spec, nil)
	return buf.String()
}

func DescribeNode(fset *token.FileSet, node ast.Node) string {
	buf := bytes.NewBuffer(make([]byte, 1024))
	ast.Fprint(buf, fset, node, nil)

	return buf.String()
}
