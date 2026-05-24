package scanner

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// TODOItem はソースコードから抽出された TODO コメントを表す。
type TODOItem struct {
	File string
	Line int
	Text string
}

var todoRe = regexp.MustCompile(`TODO:\s*(.+)`)

// ScanTODOs は rootDir 以下のソースファイルから TODO コメントを抽出する。
// .gitignore のパターンを尊重し、バイナリファイルはスキップする。
func ScanTODOs(rootDir string) ([]TODOItem, error) {
	gi := loadGitignore(rootDir)

	var items []TODOItem
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, relErr := filepath.Rel(rootDir, path)
		if relErr != nil {
			rel = path
		}

		// .git ディレクトリをスキップ
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if gi != nil && gi.MatchesPath(rel+"/") {
				return filepath.SkipDir
			}
			return nil
		}

		// .gitignore にマッチするファイルをスキップ
		if gi != nil && gi.MatchesPath(rel) {
			return nil
		}

		// テキストファイルのみ対象
		if !isTextFile(path) {
			return nil
		}

		found, scanErr := scanFile(path)
		if scanErr != nil {
			return nil
		}
		items = append(items, found...)
		return nil
	})
	return items, err
}

func loadGitignore(rootDir string) *ignore.GitIgnore {
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); err != nil {
		return nil
	}
	gi, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return nil
	}
	return gi
}

func scanFile(path string) ([]TODOItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []TODOItem
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if m := todoRe.FindStringSubmatch(line); m != nil {
			text := strings.TrimSpace(m[1])
			if text != "" {
				items = append(items, TODOItem{
					File: path,
					Line: lineNum,
					Text: text,
				})
			}
		}
	}
	return items, scanner.Err()
}

// isTextFile はファイルの最初の 512 バイトを読んでテキストファイルかどうか判定する。
func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	// 明らかなバイナリ拡張子をスキップ
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".ico": true, ".pdf": true, ".zip": true, ".tar": true,
		".gz": true, ".exe": true, ".bin": true, ".so": true,
		".db": true, ".sqlite": true, ".wasm": true,
	}
	if binaryExts[ext] {
		return false
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	for _, b := range buf[:n] {
		if b == 0 {
			return false
		}
	}
	return true
}
