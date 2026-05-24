package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/render"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "タスクをファイルに出力する",
	Args:  cobra.NoArgs,
	RunE:  runExport,
}

var exportFormat string

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "出力形式 (json|md)")
	rootCmd.AddCommand(exportCmd)
}

func runExport(_ *cobra.Command, _ []string) error {
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

	tasks, err := database.ListTasks(db.TaskFilter{
		Statuses: []db.Status{db.StatusTodo, db.StatusInProgress, db.StatusDone, db.StatusSnoozed},
	})
	if err != nil {
		r.Error(err.Error())
		return nil
	}

	switch exportFormat {
	case "json":
		return exportJSON(tasks)
	case "md", "markdown":
		return exportMarkdown(tasks)
	default:
		r.Error(fmt.Sprintf("不明な形式: %s (json|md)", exportFormat))
		return nil
	}
}

func exportJSON(tasks []db.Task) error {
	enc := json.NewEncoder(cmdOut)
	enc.SetIndent("", "  ")
	return enc.Encode(tasks)
}

func exportMarkdown(tasks []db.Task) error {
	fmt.Fprintf(cmdOut, "# Tasks\n\n_Exported: %s_\n\n", time.Now().Format("2006-01-02"))

	groups := map[db.Status][]db.Task{}
	order := []db.Status{db.StatusInProgress, db.StatusTodo, db.StatusSnoozed, db.StatusDone}
	for _, t := range tasks {
		groups[t.Status] = append(groups[t.Status], t)
	}

	labels := map[db.Status]string{
		db.StatusTodo:       "Todo",
		db.StatusInProgress: "In Progress",
		db.StatusDone:       "Done",
		db.StatusSnoozed:    "Snoozed",
	}

	for _, status := range order {
		group := groups[status]
		if len(group) == 0 {
			continue
		}
		fmt.Fprintf(cmdOut, "## %s\n\n", labels[status])
		for _, t := range group {
			due := ""
			if t.DueDate != nil {
				due = fmt.Sprintf(" _(due: %s)_", t.DueDate.Format("2006-01-02"))
			}
			tags := ""
			if len(t.Tags) > 0 {
				tags = fmt.Sprintf(" `%s`", strings.Join(t.Tags, "`, `"))
			}
			fmt.Fprintf(cmdOut, "- [%s] **#%d** %s%s%s\n",
				statusCheckbox(t.Status), t.ID, t.Title, due, tags)
		}
		fmt.Fprintln(cmdOut)
	}
	return nil
}

func statusCheckbox(s db.Status) string {
	if s == db.StatusDone {
		return "x"
	}
	return " "
}
