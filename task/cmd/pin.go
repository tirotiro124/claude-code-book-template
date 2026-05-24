package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/render"
)

var pinCmd = &cobra.Command{
	Use:   "pin <id>",
	Short: "タスクをピン固定して常に先頭表示する",
	Args:  cobra.ExactArgs(1),
	RunE:  runPin,
}

var unpinCmd = &cobra.Command{
	Use:   "unpin <id>",
	Short: "タスクのピン固定を解除する",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnpin,
}

func init() {
	rootCmd.AddCommand(pinCmd)
	rootCmd.AddCommand(unpinCmd)
}

func runPin(_ *cobra.Command, args []string) error {
	return setPinned(args[0], true)
}

func runUnpin(_ *cobra.Command, args []string) error {
	return setPinned(args[0], false)
}

func setPinned(idStr string, pinned bool) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("id は整数で指定してください: %s", idStr)
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

	task.Pinned = pinned
	if err := database.UpdateTask(task); err != nil {
		r.Error(err.Error())
		return nil
	}

	if pinned {
		fmt.Fprintf(cmdOut, "Pinned #%d: %s\n", task.ID, task.Title)
	} else {
		fmt.Fprintf(cmdOut, "Unpinned #%d: %s\n", task.ID, task.Title)
	}
	return nil
}
