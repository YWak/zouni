package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"testing"
	"unsafe"
)

type logger struct {
	t *testing.T
}

func (l *logger) Printf(fmt string, v ...any) {
	l.t.Logf(fmt, v...)
}

func (l *logger) Println(v ...any) {
	l.t.Log(v...)
}

func SetLogger(cfg interface{}, t *testing.T) {
	logger := &logger{t: t}
	pcfgType := reflect.ValueOf(cfg)
	field := pcfgType.Elem().FieldByName("log")
	field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	field.Set(reflect.ValueOf(logger))
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
