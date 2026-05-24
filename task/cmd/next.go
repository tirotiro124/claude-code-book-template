package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/git"
	"github.com/task-cli/task/internal/render"
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "次に取り組むべきタスクを表示する",
	Args:  cobra.NoArgs,
	RunE:  runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)
}

func runNext(_ *cobra.Command, _ []string) error {
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

	currentBranch := ""
	gp := git.ExecProvider{}
	if b, gErr := gp.CurrentBranch(); gErr == nil {
		currentBranch = b
	}

	if err := database.ReactivateSnoozed(time.Now()); err != nil {
		r.Error(err.Error())
		return nil
	}

	tasks, err := database.ListTasks(db.TaskFilter{})
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	if len(tasks) == 0 {
		r.NoTasks()
		return nil
	}

	scored, err := scoreAndSort(tasks, database, currentBranch, cfg.Weights)
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	top := scored[0]
	r.NextTask(top.task, top.score, currentBranch)
	_ = database.RecordShown(top.task.ID)
	return nil
}
