package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Renderer はターミナル出力を担う。
type Renderer struct {
	Stdout  io.Writer
	Stderr  io.Writer
	JSON    bool
	NoColor bool
}

// New はデフォルト設定の Renderer を返す。
func New(jsonOut, noColor bool) *Renderer {
	if noColor || os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
	return &Renderer{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		JSON:    jsonOut,
		NoColor: noColor,
	}
}

// printJSON は v を JSON として Stdout に出力する。
func (r *Renderer) printJSON(v any) {
	enc := json.NewEncoder(r.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(r.Stderr, "Error: JSON出力に失敗しました: %v\n", err)
	}
}

// colorize は NoColor が有効な場合にカラーを無効化した関数を返す。
func (r *Renderer) colorize(attrs ...color.Attribute) func(string, ...interface{}) string {
	if r.NoColor {
		return fmt.Sprintf
	}
	return color.New(attrs...).SprintfFunc()
}
