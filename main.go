package main

import (
	"fmt"
	"os"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"

	"mt2hugo/hugo" // hugoパッケージをインポート
	"mt2hugo/movabletype"
)

// ファイルシステム操作のインターフェース定義
type FileSystem interface {
	ReadFile(path string) ([]string, error)
	WriteFile(path string, content string) error
	MkdirAll(path string) error
}

// 実際のファイルシステム操作の実装
type RealFileSystem struct{}

func (fs *RealFileSystem) ReadFile(path string) ([]string, error) {
	return movabletype.ReadExportFile(path)
}

func (fs *RealFileSystem) WriteFile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func (fs *RealFileSystem) MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// コンバーター操作のインターフェース定義
type HTMLConverter interface {
	ConvertHTMLToMarkdown(html string) (string, error)
}

// 実際のコンバーター実装
type HTMLToMarkdownConverter struct {
	converter *md.Converter
}

func NewHTMLToMarkdownConverter() *HTMLToMarkdownConverter {
	converter := md.NewConverter("", true, nil)

	// 「class="keyword"」を持つリンクのカスタムルールを追加
	converter.AddRules(md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
			if selec.HasClass("keyword") {
				return &content
			}
			return nil
		},
	})

	converter.Use(plugin.GitHubFlavored())

	return &HTMLToMarkdownConverter{
		converter: converter,
	}
}

func (c *HTMLToMarkdownConverter) ConvertHTMLToMarkdown(html string) (string, error) {
	return c.converter.ConvertString(html)
}

// エントリーポイント
func main() {
	if len(os.Args) < 2 {
		fmt.Println("使い方: go run main.go <Movable_Typeエクスポートファイルのパス>")
		return
	}

	// 初期化
	fs := &RealFileSystem{}
	converter := NewHTMLToMarkdownConverter()

	// テンプレートの読み込み
	tmpl, err := hugo.LoadTemplate("templates/hugo.tmpl")
	if err != nil {
		fmt.Printf("テンプレート解析エラー: %v\n", err)
		return
	}

	// コンバーターの生成
	hugoConverter := hugo.NewConverter(fs, converter, tmpl)

	// 変換の実行
	filePath := os.Args[1]
	if err := hugoConverter.Convert(filePath, "output"); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		fmt.Println("変換が完了しました")
	}
}
