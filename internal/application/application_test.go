package application_test

import (
	"testing"

	"github.com/ywak/zouni/internal"
	"github.com/ywak/zouni/internal/application"
	"github.com/ywak/zouni/internal/config"
)

func TestApplication(t *testing.T) {
	cfg := &config.ZouniConfig{
		Debug: true,
	}
	internal.SetLogger(cfg, t)

	t.Run("parse code01", func(t *testing.T) {
		sut, err := application.Load(cfg, "../../testdata/code01")
		if err != nil {
			t.Fatalf("error: %v", err)
			return
		}

		t.Log(sut)
	})
}
