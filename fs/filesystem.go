package fs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// FileSystemError はファイルシステム操作中のエラーを表現する構造体
type FileSystemError struct {
	Op      string // 操作名 (例: "ReadFile", "WriteFile")
	Path    string // 対象のファイルパス
	Message string // エラーメッセージ
	Err     error  // 元のエラー (オプション)
}

// Error はエラーメッセージを返す
func (e *FileSystemError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ファイルシステムエラー: %s %s: %s - %v", e.Op, e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("ファイルシステムエラー: %s %s: %s", e.Op, e.Path, e.Message)
}

// Unwrap は元のエラーを返す
func (e *FileSystemError) Unwrap() error {
	return e.Err
}

// FileSystem はファイルシステム操作のインターフェース
type FileSystem interface {
	ReadFile(path string) ([]string, error)
	WriteFile(path string, content string) error
	MkdirAll(path string) error
	TryReadFile(path string) ([]string, error)
	TryWriteFile(path string, content string) error
}

// RealFileSystem は実際のファイルシステム操作を行う実装
type RealFileSystem struct{}

// ReadFile はファイルを読み込み、行のスライスを返す
func (fs *RealFileSystem) ReadFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, &FileSystemError{
			Op:      "ReadFile",
			Path:    path,
			Message: "ファイルを開けませんでした",
			Err:     err,
		}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, &FileSystemError{
			Op:      "ReadFile",
			Path:    path,
			Message: "ファイル読み込み中にエラーが発生しました",
			Err:     err,
		}
	}

	return lines, nil
}

// WriteFile はファイルに内容を書き込む
func (fs *RealFileSystem) WriteFile(path string, content string) error {
	// ディレクトリを作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return &FileSystemError{
			Op:      "WriteFile",
			Path:    dir,
			Message: "ディレクトリを作成できませんでした",
			Err:     err,
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return &FileSystemError{
			Op:      "WriteFile",
			Path:    path,
			Message: "ファイルを作成できませんでした",
			Err:     err,
		}
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return &FileSystemError{
			Op:      "WriteFile",
			Path:    path,
			Message: "ファイルに書き込めませんでした",
			Err:     err,
		}
	}

	return nil
}

// MkdirAll はディレクトリを再帰的に作成する
func (fs *RealFileSystem) MkdirAll(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return &FileSystemError{
			Op:      "MkdirAll",
			Path:    path,
			Message: "ディレクトリを作成できませんでした",
			Err:     err,
		}
	}
	return nil
}

// TryReadFile はファイルの読み込みを試みる。エラー時は空のスライスを返す
func (fs *RealFileSystem) TryReadFile(path string) ([]string, error) {
	lines, err := fs.ReadFile(path)
	if err != nil {
		return []string{}, err
	}
	return lines, nil
}

// TryWriteFile はファイルへの書き込みを試みる。エラーが発生しても処理は続行
func (fs *RealFileSystem) TryWriteFile(path string, content string) error {
	return fs.WriteFile(path, content)
}

// NewRealFileSystem は新しいRealFileSystemインスタンスを作成する
func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}
