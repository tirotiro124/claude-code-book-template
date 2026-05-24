package db

import "errors"

var (
	ErrTaskNotFound = errors.New("タスクが見つかりません")
	ErrDBSetup      = errors.New("DBの初期化に失敗しました")
)
