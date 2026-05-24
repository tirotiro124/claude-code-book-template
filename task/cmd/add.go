package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/git"
	"github.com/task-cli/task/internal/render"
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "タスクを追加する",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

var (
	addDue    string
	addBranch string
	addTags   []string
)

func init() {
	addCmd.Flags().StringVar(&addDue, "due", "", "締め切り (YYYY-MM-DD)")
	addCmd.Flags().StringVar(&addBranch, "branch", "", "関連ブランチ (省略時はカレントブランチ)")
	addCmd.Flags().StringArrayVar(&addTags, "tag", nil, "タグ (複数指定可)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	r := render.New(jsonOut, cfg.NoColor || noColor)
	r.Stdout = cmdOut
	r.Stderr = cmdErr

	database, err := db.Open()
	if err != nil {
		r.Error(err.Error())
		return nil
	}
	defer database.Close()

	branch := addBranch
	if branch == "" {
		gp := git.ExecProvider{}
		if b, gErr := gp.CurrentBranch(); gErr == nil {
			branch = b
		}
	}

	task := db.Task{
		Title:  strings.Join(args, " "),
		Status: db.StatusTodo,
		Branch: branch,
		Tags:   addTags,
	}

	if addDue != "" {
		t, pErr := time.ParseInLocation("2006-01-02", addDue, time.Local)
		if pErr != nil {
			r.Error(fmt.Sprintf("due の形式が不正です: %s (YYYY-MM-DD)", addDue))
			return nil
		}
		task.DueDate = &t
	}

	created, err := database.CreateTask(task)
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	r.Added(created)
	return nil
}
