package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/git"
	"github.com/task-cli/task/internal/render"
	"github.com/task-cli/task/internal/scanner"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "ソースコードの TODO コメントをタスクに同期する",
	Args:  cobra.NoArgs,
	RunE:  runSync,
}

var syncDryRun bool

func init() {
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "変更を加えずに結果を表示する")
	rootCmd.AddCommand(syncCmd)
}

func runSync(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	r := render.New(jsonOut, cfg.NoColor || noColor)
	r.Stdout = cmdOut
	r.Stderr = cmdErr

	gp := git.ExecProvider{}
	rootDir, err := gp.RootDir()
	if err != nil {
		r.Error("git リポジトリのルートを取得できません: " + err.Error())
		return nil
	}

	todos, err := scanner.ScanTODOs(rootDir)
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	if len(todos) == 0 {
		fmt.Fprintln(cmdOut, "TODO コメントが見つかりませんでした。")
		return nil
	}

	if syncDryRun {
		fmt.Fprintf(cmdOut, "Found %d TODO(s) (dry-run):\n", len(todos))
		for _, item := range todos {
			fmt.Fprintf(cmdOut, "  %s:%d  %s\n", item.File, item.Line, item.Text)
		}
		return nil
	}

	database, err := db.Open()
	if err != nil {
		r.Error(err.Error())
		return nil
	}
	defer database.Close()

	added := 0
	skipped := 0
	for _, item := range todos {
		// 既存タスクの重複チェック（同一 source_file + source_line）
		existing, _ := database.ListTasks(db.TaskFilter{
			Statuses: []db.Status{db.StatusTodo, db.StatusInProgress, db.StatusDone, db.StatusSnoozed},
		})
		if isDuplicate(existing, item.File, item.Line) {
			skipped++
			continue
		}

		_, err := database.CreateTask(db.Task{
			Title:      item.Text,
			SourceFile: item.File,
			SourceLine: item.Line,
		})
		if err != nil {
			r.Error(fmt.Sprintf("タスクの作成に失敗しました (%s:%d): %v", item.File, item.Line, err))
			continue
		}
		added++
		fmt.Fprintf(cmdOut, "  Added: %s\n", item.Text)
	}

	fmt.Fprintf(cmdOut, "\nSync complete: %d added, %d skipped (already exists)\n", added, skipped)
	return nil
}

func isDuplicate(tasks []db.Task, file string, line int) bool {
	for _, t := range tasks {
		if t.SourceFile == file && t.SourceLine == line {
			return true
		}
	}
	return false
}
