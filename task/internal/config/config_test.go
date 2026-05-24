package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/task-cli/task/internal/scoring"
)

func writeRC(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".taskrc")
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
	return path
}

func TestLoad_Default(t *testing.T) {
	// .taskrc が存在しないディレクトリで実行
	tmp := t.TempDir()
	t.Chdir(tmp)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetBalanced, cfg.Preset)
	assert.Equal(t, scoring.PresetWeights[scoring.PresetBalanced], cfg.Weights)
}

func TestLoad_DeadlinePreset(t *testing.T) {
	tmp := t.TempDir()
	writeRC(t, tmp, `preset = "deadline"`)
	t.Chdir(tmp)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetDeadline, cfg.Preset)
	assert.Equal(t, scoring.PresetWeights[scoring.PresetDeadline], cfg.Weights)
}

func TestLoad_FlowPreset(t *testing.T) {
	tmp := t.TempDir()
	writeRC(t, tmp, `preset = "flow"`)
	t.Chdir(tmp)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetFlow, cfg.Preset)
}

func TestLoad_InvalidPreset(t *testing.T) {
	tmp := t.TempDir()
	writeRC(t, tmp, `preset = "invalid"`)
	t.Chdir(tmp)

	_, err := Load()
	assert.ErrorIs(t, err, ErrInvalidPreset)
}

func TestLoad_EmptyPreset(t *testing.T) {
	tmp := t.TempDir()
	writeRC(t, tmp, `preset = ""`)
	t.Chdir(tmp)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetBalanced, cfg.Preset)
}

func TestLoad_HomeFallback(t *testing.T) {
	// カレントに .taskrc がなく HOME にある場合
	homeDir := t.TempDir()
	workDir := t.TempDir()
	writeRC(t, homeDir, `preset = "flow"`)
	t.Setenv("HOME", homeDir)
	t.Chdir(workDir)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetFlow, cfg.Preset)
}

func TestLoad_NoColor(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	t.Setenv("NO_COLOR", "1")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.NoColor)
}

func TestLoad_ParentDir(t *testing.T) {
	// 親ディレクトリの .taskrc を発見するケース
	parentDir := t.TempDir()
	childDir := filepath.Join(parentDir, "sub")
	require.NoError(t, os.MkdirAll(childDir, 0700))
	writeRC(t, parentDir, `preset = "deadline"`)
	t.Chdir(childDir)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, scoring.PresetDeadline, cfg.Preset)
}
