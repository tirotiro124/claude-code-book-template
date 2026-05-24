package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/render"
)

var snoozeCmd = &cobra.Command{
	Use:   "snooze <id> <duration>",
	Short: "タスクを一時的に非表示にする (例: 1d / 2h / 1w)",
	Args:  cobra.ExactArgs(2),
	RunE:  runSnooze,
}

func init() {
	rootCmd.AddCommand(snoozeCmd)
}

func runSnooze(_ *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("id は整数で指定してください: %s", args[0])
	}

	d, err := parseSnoozeDuration(args[1])
	if err != nil {
		return err
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

	until := time.Now().Add(d)
	task.Status = db.StatusSnoozed
	task.SnoozedUntil = &until
	if err := database.UpdateTask(task); err != nil {
		r.Error(err.Error())
		return nil
	}

	fmt.Fprintf(cmdOut, "Snoozed #%d until %s\n", task.ID, until.Format("2006-01-02"))
	return nil
}

// parseSnoozeDuration は "1d" "2h" "1w" などのショートハンドを含む期間文字列を解析する。
func parseSnoozeDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("期間の形式が不正です: %s (例: 1d, 2h, 1w)", s)
	}
	suffix := s[len(s)-1]
	num, err := strconv.Atoi(s[:len(s)-1])
	if err != nil || num <= 0 {
		// Go 標準の duration 形式にフォールバック
		return time.ParseDuration(s)
	}
	switch suffix {
	case 'd':
		return time.Duration(num) * 24 * time.Hour, nil
	case 'w':
		return time.Duration(num) * 7 * 24 * time.Hour, nil
	case 'h':
		return time.Duration(num) * time.Hour, nil
	case 'm':
		return time.Duration(num) * time.Minute, nil
	default:
		return 0, fmt.Errorf("期間の単位が不正です: %c (d/h/m/w)", suffix)
	}
}
