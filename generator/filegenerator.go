package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"mt2hugo/downloader" // 追加
	"mt2hugo/fs"
	"mt2hugo/models" // hugoパッケージからmodelsパッケージに変更
	"mt2hugo/util"
)

// FileGenerator はHugo記事ファイルの生成を担当する
type FileGenerator struct {
	fs              fs.FileSystem
	tmpl            *template.Template
	baseDir         string
	imageDownloader *downloader.ImageDownloader // 追加
}

// NewFileGenerator は新しいFileGeneratorを作成する
func NewFileGenerator(
	fileSystem fs.FileSystem,
	tmpl *template.Template,
	baseDir string,
	imageDownloader *downloader.ImageDownloader, // 追加
) *FileGenerator {
	return &FileGenerator{
		fs:              fileSystem,
		tmpl:            tmpl,
		baseDir:         baseDir,
		imageDownloader: imageDownloader,
	}
}

// GenerateFile はHugo記事ファイルを生成する
// GenerateFile はHugo記事ファイルを生成する
func (g *FileGenerator) GenerateFile(article models.HugoArticle, dateTime time.Time) (string, error) {
	// 出力ディレクトリパスを生成
	dirName := util.FormatDirName(dateTime)
	dirPath := filepath.Join(g.baseDir, dirName)

	// ディレクトリを作成
	if err := g.fs.MkdirAll(dirPath); err != nil {
		return "", fmt.Errorf("ディレクトリ作成エラー: %w", err)
	}

	// 記事処理のコピーを作成
	processedArticle := article

	// 画像ダウンロード処理
	if g.imageDownloader != nil {
		// 本文内の画像処理
		if article.Body != "" {
			processedBody, err := g.imageDownloader.ProcessHTMLImages(article.Body, dirPath)
			if err != nil {
				fmt.Printf("警告: 本文の画像ダウンロード中にエラー: %v\n", err)
			} else {
				processedArticle.Body = processedBody
			}
		}

		// 拡張本文の画像処理
		if article.ExtendedBody != "" {
			processedExtBody, err := g.imageDownloader.ProcessHTMLImages(article.ExtendedBody, dirPath)
			if err != nil {
				fmt.Printf("警告: 拡張本文の画像ダウンロード中にエラー: %v\n", err)
			} else {
				processedArticle.ExtendedBody = processedExtBody
			}
		}

		// IMAGE フィールドの画像処理
		if article.Image != "" {
			// IMAGE フィールドの URL を処理
			tempContent := fmt.Sprintf("IMAGE: %s", article.Image)
			processedContent, err := g.imageDownloader.ProcessHTMLImages(tempContent, dirPath)
			if err != nil {
				fmt.Printf("警告: IMAGE フィールドの画像ダウンロード中にエラー: %v\n", err)
			} else {
				// "IMAGE: " を除去して実際のファイル名だけを取得
				processedArticle.Image = strings.TrimPrefix(processedContent, "IMAGE: ")
			}
		}
	}

	// テンプレートを使って出力内容を生成
	var output strings.Builder
	if err := g.tmpl.Execute(&output, processedArticle); err != nil {
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
