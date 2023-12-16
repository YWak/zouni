package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Logger interface {
	Printf(fmt string, v ...any)
	Println(v ...any)
}

type NopLogger struct{}

func (log *NopLogger) Printf(fmt string, v ...any) {
	// NOP
}

func (log *NopLogger) Println(v ...any) {
	// NOP
}

var _ Logger = &NopLogger{}

type ZouniConfig struct {
	Excludes []string `json:"excludes"`
	Debug    bool     `json:"debug"`
	log      Logger
}

// NewZouniConfigはコンフィグを返します。
func NewZouniConfig(targetFile string) (*ZouniConfig, error) {
	cfg := ZouniConfig{
		Excludes: []string{},
		Debug:    false,
	}

	err := cfg.Load(targetFile)
	if err != nil {
		return nil, err
	}
	if cfg.Debug {
		cfg.log = log.Default()
	} else {
		cfg.log = &NopLogger{}
	}

	return &cfg, nil
}

func exists(dirName string) bool {
	_, err := os.Stat(dirName)
	return err != nil
}

// Loadは設定情報を読み込みます。
func (cfg *ZouniConfig) Load(targetFile string) error {
	abspath, err := filepath.Abs(targetFile)
	if err != nil {
		return err
	}

	for dir := filepath.Dir(abspath); dir != "/"; dir = filepath.Dir(dir) {
		if !exists(dir) {
			break
		}
		file := filepath.Join(dir, ".zounirc")
		if exists(file) {
			continue
		}
		err = cfg.loadFrom(file)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// loadFromはファイルを解析して結果を保存します。
func (cfg *ZouniConfig) loadFrom(file string) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *ZouniConfig) Logger() Logger {
	return cfg.log
}
