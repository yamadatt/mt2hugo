package fs

import (
	"os"

	"mt2hugo/movabletype"
)

// FileSystem はファイルシステム操作のインターフェース
type FileSystem interface {
	ReadFile(path string) ([]string, error)
	WriteFile(path string, content string) error
	MkdirAll(path string) error
}

// RealFileSystem は実際のファイルシステム操作を行う実装
type RealFileSystem struct{}

// ReadFile はファイルを読み込み、行のスライスを返す
func (fs *RealFileSystem) ReadFile(path string) ([]string, error) {
	return movabletype.ReadExportFile(path)
}

// WriteFile はファイルに内容を書き込む
func (fs *RealFileSystem) WriteFile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// MkdirAll はディレクトリを再帰的に作成する
func (fs *RealFileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// NewRealFileSystem は新しいRealFileSystemインスタンスを作成する
func NewRealFileSystem() *RealFileSystem {
	return &RealFileSystem{}
}
