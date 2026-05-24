package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/task-cli/task/internal/config"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/render"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "$EDITOR でタスクをTOML形式で編集する",
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

// editableTask は TOML 編集用のフラット構造体。
type editableTask struct {
	Title  string `toml:"title"`
	Due    string `toml:"due"` // YYYY-MM-DD or ""
	Branch string `toml:"branch"`
	Tags   string `toml:"tags"`   // カンマ区切り
	Status string `toml:"status"` // todo|in_progress|done|snoozed
}

func runEdit(_ *cobra.Command, args []string) error {
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("id は整数で指定してください: %s", args[0])
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return fmt.Errorf("$EDITOR が設定されていません。例: export EDITOR=vim")
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

	// タスクを TOML に変換して一時ファイルに書き出す
	et := taskToEditable(task)
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("task-%d-*.toml", id))
	if err != nil {
		r.Error(fmt.Sprintf("一時ファイルの作成に失敗しました: %v", err))
		return nil
	}
	defer os.Remove(tmpFile.Name())

	enc := toml.NewEncoder(tmpFile)
	if err := enc.Encode(et); err != nil {
		r.Error(fmt.Sprintf("TOML のエンコードに失敗しました: %v", err))
		return nil
	}
	tmpFile.Close()

	// エディタを起動
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		r.Error(fmt.Sprintf("エディタの起動に失敗しました: %v", err))
		return nil
	}

	// 編集後の TOML を読み込む
	var updated editableTask
	if _, err := toml.DecodeFile(tmpFile.Name(), &updated); err != nil {
		r.Error(fmt.Sprintf("TOML のデコードに失敗しました: %v", err))
		return nil
	}

	// Task に反映
	if err := applyEditable(&task, updated); err != nil {
		r.Error(err.Error())
		return nil
	}

	if err := database.UpdateTask(task); err != nil {
		r.Error(err.Error())
		return nil
	}

	fmt.Fprintf(cmdOut, "Updated #%d: %s\n", task.ID, task.Title)
	return nil
}

func taskToEditable(t db.Task) editableTask {
	due := ""
	if t.DueDate != nil {
		due = t.DueDate.Format("2006-01-02")
	}
	return editableTask{
		Title:  t.Title,
		Due:    due,
		Branch: t.Branch,
		Tags:   strings.Join(t.Tags, ", "),
		Status: string(t.Status),
	}
}

func applyEditable(t *db.Task, e editableTask) error {
	t.Title = strings.TrimSpace(e.Title)
	if t.Title == "" {
		return fmt.Errorf("title は空にできません")
	}

	if e.Due == "" {
		t.DueDate = nil
	} else {
		parsed, err := time.ParseInLocation("2006-01-02", e.Due, time.Local)
		if err != nil {
			return fmt.Errorf("due の形式が不正です: %s (YYYY-MM-DD)", e.Due)
		}
		t.DueDate = &parsed
	}

	t.Branch = strings.TrimSpace(e.Branch)

	tags := []string{}
	for _, tag := range strings.Split(e.Tags, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	t.Tags = tags

	switch db.Status(e.Status) {
	case db.StatusTodo, db.StatusInProgress, db.StatusDone, db.StatusSnoozed:
		t.Status = db.Status(e.Status)
	default:
		return fmt.Errorf("status が不正です: %s", e.Status)
	}

	return nil
}
