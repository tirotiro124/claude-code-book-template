package render

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/task-cli/task/internal/db"
)

// Added は task add 完了時のメッセージを表示する。
func (r *Renderer) Added(task db.Task) {
	if r.JSON {
		r.printJSON(task)
		return
	}
	green := r.colorize(color.FgGreen)
	msg := fmt.Sprintf("Added #%d: %s", task.ID, task.Title)
	if task.Branch != "" {
		msg += fmt.Sprintf("  [branch: %s]", task.Branch)
	}
	fmt.Fprintln(r.Stdout, green("%s", msg))
}

// Done は task done 完了時のメッセージを表示する。
func (r *Renderer) Done(task db.Task) {
	if r.JSON {
		r.printJSON(task)
		return
	}
	green := r.colorize(color.FgGreen)
	fmt.Fprintf(r.Stdout, "%s\n\n", green("Done #%d: %s", task.ID, task.Title))
}

// NoTasks はタスクが0件のときのメッセージを表示する。
func (r *Renderer) NoTasks() {
	fmt.Fprintln(r.Stdout, "No tasks found.")
	fmt.Fprintln(r.Stdout, `  Add a task: task add "タスク内容"`)
	fmt.Fprintln(r.Stdout, "  Import TODOs: task sync")
}

// Error はエラーメッセージを stderr に表示する。
func (r *Renderer) Error(msg string) {
	red := r.colorize(color.FgRed)
	fmt.Fprintln(r.Stderr, red("Error: %s", msg))
}
