package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/git"
	"github.com/task-cli/task/internal/render"
	"github.com/task-cli/task/internal/scoring"
)

var showCmd = &cobra.Command{
	Use:     "show <id>",
	Aliases: []string{"info"},
	Short:   "タスクの詳細を表示する",
	Args:    cobra.ExactArgs(1),
	RunE:    runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(_ *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("id は整数で指定してください: %s", args[0])
	}

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

	task, err := database.GetTask(id)
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	currentBranch := ""
	gp := git.ExecProvider{}
	if b, gErr := gp.CurrentBranch(); gErr == nil {
		currentBranch = b
	}

	recentShown, _ := database.GetRecentShown()
	cnt, _ := database.BlockedByCount(task.ID)
	prereqs, _ := database.GetPrerequisites(task.ID)

	in := scoring.Input{
		TaskID:         task.ID,
		DueDate:        task.DueDate,
		BlockedByCount: cnt,
		Branch:         task.Branch,
		CurrentBranch:  currentBranch,
		CreatedAt:      task.CreatedAt,
		Now:            time.Now(),
		RecentShown:    recentShown,
		Weights:        cfg.Weights,
	}
	score := scoring.Calculate(in)

	r.TaskDetail(task, score, prereqs)
	return nil
}
