package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/scoring"
)

// NextTask は task next の出力を表示する。
func (r *Renderer) NextTask(task db.Task, score scoring.ScoreResult, currentBranch string) {
	if r.JSON {
		r.printJSON(map[string]any{
			"task":  task,
			"score": score,
		})
		return
	}

	red := r.colorize(color.FgRed)
	orange := r.colorize(color.FgYellow)
	bold := r.colorize(color.Bold)
	dim := r.colorize(color.Faint)

	fmt.Fprintf(r.Stdout, "┌─ Next Task %s\n", strings.Repeat("─", 38))
	fmt.Fprintf(r.Stdout, "│ %s  %s\n", bold("#%d", task.ID), bold("%s", task.Title))
	fmt.Fprintln(r.Stdout, "│")

	// スコア行
	dueStr := formatDue(task.DueDate, time.Now(), red, orange)
	branchStr := ""
	if task.Branch != "" && task.Branch == currentBranch {
		branchStr = "  branch: ✓"
	}
	blocksStr := ""
	if score.Importance > 0 {
		blocksStr = fmt.Sprintf("  blocks: %d", score.Importance/10)
	}
	fmt.Fprintf(r.Stdout, "│  Score %s  |%s%s%s\n",
		bold("%d", score.Total), dueStr, blocksStr, branchStr)

	// スコア内訳行
	fmt.Fprintf(r.Stdout, "│  %s\n", dim(
		"urgency:%d  deps:%d  ctx:%d  age:%d  fatigue:-%d",
		score.Urgency, score.Importance, score.Context, score.Aging, score.Fatigue,
	))

	fmt.Fprintln(r.Stdout, "│")
	fmt.Fprintf(r.Stdout, "│  %s     %s\n",
		dim("$ task done %d", task.ID),
		dim("$ task snooze %d 1d", task.ID),
	)
	fmt.Fprintf(r.Stdout, "└%s\n", strings.Repeat("─", 51))
}

// TaskList は task list の出力を表示する。
func (r *Renderer) TaskList(tasks []db.Task, scores []scoring.ScoreResult) {
	if r.JSON {
		rows := make([]map[string]any, len(tasks))
		for i, t := range tasks {
			rows[i] = map[string]any{"task": t, "score": scores[i]}
		}
		r.printJSON(rows)
		return
	}

	if len(tasks) == 0 {
		r.NoTasks()
		return
	}

	red := r.colorize(color.FgRed)
	orange := r.colorize(color.FgYellow)

	t := table.NewWriter()
	t.SetOutputMirror(r.Stdout)
	t.AppendHeader(table.Row{"#", "score", "title", "due", "branch", "tags"})

	for i, task := range tasks {
		dueStr := formatDuePlain(task.DueDate, time.Now(), red, orange)
		tags := strings.Join(task.Tags, ",")
		t.AppendRow(table.Row{
			fmt.Sprintf("%d", task.ID),
			fmt.Sprintf("%d", scores[i].Total),
			task.Title,
			dueStr,
			task.Branch,
			tags,
		})
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}

// TaskDetail は task show の出力を表示する。
func (r *Renderer) TaskDetail(task db.Task, score scoring.ScoreResult, blockedBy []db.Task) {
	if r.JSON {
		r.printJSON(map[string]any{
			"task":       task,
			"score":      score,
			"blocked_by": blockedBy,
		})
		return
	}

	bold := r.colorize(color.Bold)
	dim := r.colorize(color.Faint)
	sep := strings.Repeat("─", 44)

	fmt.Fprintf(r.Stdout, "%s  %s\n", bold("#%d", task.ID), bold("%s", task.Title))
	fmt.Fprintln(r.Stdout, sep)
	fmt.Fprintf(r.Stdout, "status    : %s\n", task.Status)
	fmt.Fprintf(r.Stdout, "score     : %d\n", score.Total)
	fmt.Fprintf(r.Stdout, "  urgency : %s\n", dim("%d  (%s)", score.Urgency, dueDesc(task.DueDate, time.Now())))
	fmt.Fprintf(r.Stdout, "  deps    : %s\n", dim("%d  (%d件がこのタスク待ち)", score.Importance, len(blockedBy)))
	fmt.Fprintf(r.Stdout, "  context : %s\n", dim("%d", score.Context))
	fmt.Fprintf(r.Stdout, "  aging   : %s\n", dim("%d  (作成から%d日未着手)", score.Aging, int(time.Since(task.CreatedAt).Hours()/24)))
	fmt.Fprintf(r.Stdout, "  fatigue : %s\n", dim("-%d", score.Fatigue))

	if task.DueDate != nil {
		fmt.Fprintf(r.Stdout, "due       : %s\n", task.DueDate.Format("2006-01-02"))
	}
	if task.Branch != "" {
		fmt.Fprintf(r.Stdout, "branch    : %s\n", task.Branch)
	}
	if len(task.Tags) > 0 {
		fmt.Fprintf(r.Stdout, "tags      : %s\n", strings.Join(task.Tags, ", "))
	}
	if len(blockedBy) > 0 {
		ids := make([]string, len(blockedBy))
		for i, b := range blockedBy {
			ids[i] = fmt.Sprintf("#%d", b.ID)
		}
		fmt.Fprintf(r.Stdout, "blocks    : %s\n", strings.Join(ids, ", "))
	}
	fmt.Fprintln(r.Stdout, sep)
}

// --- ヘルパー ---

func formatDue(due *time.Time, now time.Time, red, orange func(string, ...interface{}) string) string {
	if due == nil {
		return ""
	}
	days := int(due.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	switch {
	case days < 0:
		return fmt.Sprintf("  due: %s", red("期限切れ%dd", -days))
	case days == 0:
		return fmt.Sprintf("  due: %s", red("当日"))
	case days <= 3:
		return fmt.Sprintf("  due: %s", orange("%dd", days))
	default:
		return fmt.Sprintf("  due: %dd", days)
	}
}

func formatDuePlain(due *time.Time, now time.Time, red, orange func(string, ...interface{}) string) string {
	if due == nil {
		return "-"
	}
	days := int(due.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	switch {
	case days < 0:
		return red("期限切れ")
	case days == 0:
		return red("当日")
	case days <= 3:
		return orange("%dd", days)
	default:
		return fmt.Sprintf("%dd", days)
	}
}

func dueDesc(due *time.Time, now time.Time) string {
	if due == nil {
		return "締め切りなし"
	}
	days := int(due.Truncate(24*time.Hour).Sub(now.Truncate(24*time.Hour)).Hours() / 24)
	if days < 0 {
		return fmt.Sprintf("%d日超過", -days)
	}
	return fmt.Sprintf("締め切りまで%d日", days)
}
