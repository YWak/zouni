package internal

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ZouniConfig struct {
	Excludes []string `json:"excludes"`
	Debug    bool     `json:"debug"`
}

var Config *ZouniConfig = &ZouniConfig{}

func (cfg *ZouniConfig) Load(dir string) error {
	// set default value
	cfg.Excludes = []string{}
	cfg.Debug = false

	err := cfg.load(dir)
	if err != nil {
		return err
	}

	if cfg.Debug {
		Log = log.Default()
	} else {
		Log = &NopLogger{}
	}

	return nil
}

func exists(dirName string) bool {
	_, err := os.Stat(dirName)
	return err != nil
}

// Loadは設定情報を読み込みます。
func (cfg *ZouniConfig) load(directory string) error {
	abspath, err := filepath.Abs(directory)
	if err != nil {
		return err
	}

	for dir := abspath; dir != "/"; dir = filepath.Dir(dir) {
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
