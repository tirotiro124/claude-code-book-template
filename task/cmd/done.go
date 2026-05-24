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
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "タスクを完了にする",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(_ *cobra.Command, args []string) error {
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

	task.Status = db.StatusDone
	task.UpdatedAt = time.Now()
	if err := database.UpdateTask(task); err != nil {
		r.Error(err.Error())
		return nil
	}

	if err := database.RemoveAllBlocksBy(id); err != nil {
		r.Error(err.Error())
		return nil
	}

	r.Done(task)

	// Show next task after completion
	currentBranch := ""
	gp := git.ExecProvider{}
	if b, gErr := gp.CurrentBranch(); gErr == nil {
		currentBranch = b
	}

	_ = database.ReactivateSnoozed(time.Now())
	tasks, err := database.ListTasks(db.TaskFilter{})
	if err != nil || len(tasks) == 0 {
		return nil
	}

	scored, err := scoreAndSort(tasks, database, currentBranch, cfg.Weights)
	if err != nil || len(scored) == 0 {
		return nil
	}

	top := scored[0]
	r.NextTask(top.task, top.score, currentBranch)
	_ = database.RecordShown(top.task.ID)
	return nil
}
