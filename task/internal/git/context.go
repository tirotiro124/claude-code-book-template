package git

import (
	"errors"
	"os/exec"
	"strings"
)

// ErrNotGitRepository は git リポジトリ外での実行を表す sentinel error。
var ErrNotGitRepository = errors.New("git リポジトリ外での実行です")

// Provider は git コンテキストへのアクセスを抽象化するインターフェース。
type Provider interface {
	CurrentBranch() (string, error)
	RootDir() (string, error)
}

// ExecProvider は git バイナリを外部プロセスとして呼ぶ本番実装。
type ExecProvider struct {
	// Dir を指定するとそのディレクトリで git コマンドを実行する（省略時はカレント）
	Dir string
}

// CurrentBranch は現在のブランチ名を返す。
// git リポジトリ外の場合は ("", ErrNotGitRepository) を返す。
func (p ExecProvider) CurrentBranch() (string, error) {
	out, err := p.run("branch", "--show-current")
	if err != nil {
		return "", ErrNotGitRepository
	}
	branch := strings.TrimSpace(out)
	if branch == "" {
		return "", ErrNotGitRepository
	}
	return branch, nil
}

// RootDir は git リポジトリのルートディレクトリを返す。
func (p ExecProvider) RootDir() (string, error) {
	out, err := p.run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", ErrNotGitRepository
	}
	return strings.TrimSpace(out), nil
}

func (p ExecProvider) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if p.Dir != "" {
		cmd.Dir = p.Dir
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
