package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/render"
)

var blockCmd = &cobra.Command{
	Use:   "block <id> <blocked-by-id>",
	Short: "タスクに依存関係を追加する（id は blocked-by-id の完了を待つ）",
	Args:  cobra.ExactArgs(2),
	RunE:  runBlock,
}

var unblockCmd = &cobra.Command{
	Use:   "unblock <id> <blocked-by-id>",
	Short: "タスクの依存関係を解除する",
	Args:  cobra.ExactArgs(2),
	RunE:  runUnblock,
}

func init() {
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(unblockCmd)
}

func runBlock(_ *cobra.Command, args []string) error {
	taskID, blockedByID, r, database, err := parseBlockArgs(args)
	if err != nil {
		return err
	}
	if database == nil {
		return nil
	}
	defer database.Close()

	if taskID == blockedByID {
		r.Error("タスクは自分自身をブロックできません")
		return nil
	}

	// 両タスクの存在確認
	if _, err := database.GetTask(taskID); err != nil {
		r.Error(err.Error())
		return nil
	}
	blockedBy, err := database.GetTask(blockedByID)
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	if err := database.AddBlock(taskID, blockedByID); err != nil {
		r.Error(err.Error())
		return nil
	}

	fmt.Fprintf(cmdOut, "Block added: #%d now waits for #%d (%s)\n", taskID, blockedBy.ID, blockedBy.Title)
	return nil
}

func runUnblock(_ *cobra.Command, args []string) error {
	taskID, blockedByID, r, database, err := parseBlockArgs(args)
	if err != nil {
		return err
	}
	if database == nil {
		return nil
	}
	defer database.Close()

	if err := database.RemoveBlock(taskID, blockedByID); err != nil {
		r.Error(err.Error())
		return nil
	}

	fmt.Fprintf(cmdOut, "Block removed: #%d no longer waits for #%d\n", taskID, blockedByID)
	return nil
}

func parseBlockArgs(args []string) (int64, int64, *render.Renderer, *db.DB, error) {
	taskID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("id は整数で指定してください: %s", args[0])
	}
	blockedByID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return 0, 0, nil, nil, fmt.Errorf("id は整数で指定してください: %s", args[1])
	}

	cfg, err := config.Load()
	if err != nil {
		return 0, 0, nil, nil, err
	}
	r := render.New(jsonOut, cfg.NoColor || noColor)
	r.Stdout = cmdOut
	r.Stderr = cmdErr

	database, err := db.Open()
	if err != nil {
		r.Error(err.Error())
		return 0, 0, r, nil, nil
	}

	return taskID, blockedByID, r, database, nil
}
