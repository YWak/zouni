package application_test

import (
	"bytes"
	"fmt"
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

		decls, err := sut.Dig()
		if err != nil {
			t.Fatalf("error on Dig(): %v", err)
		}

		buf := bytes.NewBuffer([]byte{})
		pcfg := printer.Config{
			Mode:     printer.TabIndent | printer.SourcePos,
			Tabwidth: 4,
		}
		for _, decl := range decls {
			pcfg.Fprint(buf, fset, decl)
			fmt.Fprintf(buf, "\n\n")
		}

		t.Logf("dig =>\n%s", buf.String())
	})
}
