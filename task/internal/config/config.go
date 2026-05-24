package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/task-cli/task/internal/scoring"
)

// ErrInvalidPreset は不正なプリセット値を表す sentinel error。
var ErrInvalidPreset = errors.New("不正なプリセット値です（deadline / balanced / flow のいずれかを指定してください）")

// Config はアプリ全体の設定を保持する。
type Config struct {
	Preset  scoring.Preset
	Weights scoring.WeightMap
	NoColor bool
}

// taskrc はTOMLファイルの構造体。
type taskrc struct {
	Preset string `toml:"preset"`
}

// Load は .taskrc を探索して Config を返す。
// カレントディレクトリから上位へ探索し、見つからなければ ~/.taskrc を確認する。
// どちらも存在しない場合はデフォルト値（PresetBalanced）を返す。
func Load() (Config, error) {
	path, found := findRC()
	if !found {
		return defaultConfig(), nil
	}
	return loadFrom(path)
}

// loadFrom は指定パスの .taskrc を読み込む。
func loadFrom(path string) (Config, error) {
	var rc taskrc
	if _, err := toml.DecodeFile(path, &rc); err != nil {
		return Config{}, err
	}

	if rc.Preset == "" {
		return defaultConfig(), nil
	}

	preset := scoring.Preset(rc.Preset)
	weights, ok := scoring.PresetWeights[preset]
	if !ok {
		return Config{}, ErrInvalidPreset
	}

	return Config{
		Preset:  preset,
		Weights: weights,
		NoColor: isNoColor(),
	}, nil
}

// findRC は .taskrc ファイルのパスを探索して返す。
func findRC() (string, bool) {
	// カレントディレクトリから上位へ探索
	dir, err := os.Getwd()
	if err == nil {
		for {
			candidate := filepath.Join(dir, ".taskrc")
			if _, err := os.Stat(candidate); err == nil {
				return candidate, true
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// ~/.taskrc へのフォールバック
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, ".taskrc")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}

	return "", false
}

func defaultConfig() Config {
	return Config{
		Preset:  scoring.PresetBalanced,
		Weights: scoring.PresetWeights[scoring.PresetBalanced],
		NoColor: isNoColor(),
	}
}

func isNoColor() bool {
	return os.Getenv("NO_COLOR") != ""
}
