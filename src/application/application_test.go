package application_test

import (
	"testing"

	"github.com/ywak/zouni/src"
	"github.com/ywak/zouni/src/application"
	"github.com/ywak/zouni/src/config"
)

func TestApplication(t *testing.T) {
	cfg := &config.ZouniConfig{
		Debug: true,
	}
	src.SetLogger(cfg, t)

	t.Run("parse code01", func(t *testing.T) {
		sut, err := application.Load(cfg, "../../testdata/code01")
		if err != nil {
			t.Fatalf("error: %v", err)
			return
		}

		t.Log(sut)
	})
}
