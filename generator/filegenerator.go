package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"mt2hugo/fs"
	"mt2hugo/models" // hugoパッケージからmodelsパッケージに変更
	"mt2hugo/util"
)

// FileGenerator はHugo記事ファイルの生成を担当する
type FileGenerator struct {
	fs      fs.FileSystem
	tmpl    *template.Template
	baseDir string
}

// NewFileGenerator は新しいFileGeneratorを作成する
func NewFileGenerator(fileSystem fs.FileSystem, tmpl *template.Template, baseDir string) *FileGenerator {
	return &FileGenerator{
		fs:      fileSystem,
		tmpl:    tmpl,
		baseDir: baseDir,
	}
}

// GenerateFile はHugo記事ファイルを生成する
func (g *FileGenerator) GenerateFile(article models.HugoArticle, dateTime time.Time) (string, error) {
	// 以下、既存のコード
	// 出力ディレクトリパスを生成
	dirName := util.FormatDirName(dateTime)
	dirPath := filepath.Join(g.baseDir, dirName)

	// ディレクトリを作成
	if err := g.fs.MkdirAll(dirPath); err != nil {
		return "", fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	// テンプレートを使って出力内容を生成
	var output strings.Builder
	if err := g.tmpl.Execute(&output, article); err != nil {
		return "", fmt.Errorf("テンプレート実行エラー: %w", err)
	}

	// ファイルに書き込み
	filePath := filepath.Join(dirPath, "index.md")
	if err := g.fs.WriteFile(filePath, output.String()); err != nil {
		return "", fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	return filePath, nil
}

// SetBaseDir はベースディレクトリを設定する
func (g *FileGenerator) SetBaseDir(baseDir string) {
	g.baseDir = baseDir
}

// GetBaseDir は現在のベースディレクトリを取得する
func (g *FileGenerator) GetBaseDir() string {
	return g.baseDir
}
