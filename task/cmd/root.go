package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	jsonOut bool
	noColor bool

	cmdOut io.Writer = os.Stdout
	cmdErr io.Writer = os.Stderr
)

var rootCmd = &cobra.Command{
	Use:   "task",
	Short: "開発者向けCLIタスク管理ツール",
	Long:  `Gitコンテキストに応じて優先順位を自動調整するCLIタスク管理ツール。`,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "スコア計算の詳細ログを表示する")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "JSON形式で出力する")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "カラー出力を無効にする")
}
