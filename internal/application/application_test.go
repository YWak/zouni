package application_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"testing"

	"github.com/ywak/zouni/internal"
	"github.com/ywak/zouni/internal/application"
)

func TestApplication(t *testing.T) {
	internal.Log = &internal.TestLogger{T: t}
	t.Run("parse code01", func(t *testing.T) {
		fset := token.NewFileSet()
		sut, err := application.Load(fset, "../../testdata/code01")
		if err != nil {
			t.Fatalf("error: %v", err)
			return
		}

		for _, decl := range sut.Decls() {
			buf := new(bytes.Buffer)
			ast.Fprint(buf, fset, decl, nil)
			t.Logf("decl =>\n%s\n", buf.String())
		}

		decls, err := sut.Dig()
		if err != nil {
			t.Fatalf("error on Dig(): %v", err)
		}

		buf := new(bytes.Buffer)
		pcfg := printer.Config{
			Mode:     printer.TabIndent | printer.SourcePos,
			Tabwidth: 4,
		}
		for _, decl := range decls {
			pcfg.Fprint(buf, fset, decl)
			fmt.Fprintf(buf, "\n\n")
		}

		t.Logf("dig =>\n%s\n", buf.String())
		for _, decl := range decls {
			buf := new(bytes.Buffer)
			ast.Fprint(buf, fset, decl, nil)
			// t.Logf(buf.String())
		}
	})
}
