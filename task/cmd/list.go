package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/git"
	"github.com/task-cli/task/internal/render"
	"github.com/task-cli/task/internal/scoring"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "タスク一覧を表示する",
	Args:    cobra.NoArgs,
	RunE:    runList,
}

var (
	listStatus string
	listTag    string
	listAll    bool
)

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "ステータスでフィルタ (todo|in_progress|done|snoozed)")
	listCmd.Flags().StringVar(&listTag, "tag", "", "タグでフィルタ")
	listCmd.Flags().BoolVar(&listAll, "all", false, "完了・スヌーズ含む全タスクを表示")
	rootCmd.AddCommand(listCmd)
}

func runList(_ *cobra.Command, _ []string) error {
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

	filter := db.TaskFilter{Tag: listTag}
	switch {
	case listAll:
		filter.Statuses = []db.Status{db.StatusTodo, db.StatusInProgress, db.StatusDone, db.StatusSnoozed}
	case listStatus != "":
		filter.Statuses = []db.Status{db.Status(listStatus)}
	}

	tasks, err := database.ListTasks(filter)
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

	finalTasks := make([]db.Task, len(scored))
	finalScores := make([]scoring.ScoreResult, len(scored))
	for i, s := range scored {
		finalTasks[i] = s.task
		finalScores[i] = s.score
	}
	r.TaskList(finalTasks, finalScores)
	return nil
}
