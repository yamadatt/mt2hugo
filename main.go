package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"

	"mt2hugo/movabletype" // パッケージのインポート
	"mt2hugo/util"        // utilパッケージを追加
)

// 記事データを保持する構造体
type HugoArticle struct {
	Title    string
	Date     string
	Slug     string
	Category string
	Tags     []string
	Image    string
	Summary  string
	Content  string
}

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
type Converter interface {
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

// アプリケーションのメイン処理を担当する構造体
type HugoConverter struct {
	fs        FileSystem
	converter Converter
	tmpl      *template.Template
}

func NewHugoConverter(fs FileSystem, converter Converter, tmpl *template.Template) *HugoConverter {
	return &HugoConverter{
		fs:        fs,
		converter: converter,
		tmpl:      tmpl,
	}
}

// ファイル変換のメイン処理
func (h *HugoConverter) Convert(inputPath string, outputBaseDir string) error {
	// ファイル読み込み
	lines, err := h.fs.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("エクスポートファイル読み込みエラー: %v", err)
	}

	// パース処理
	articleMaps := movabletype.ParseExportFile(lines)
	articles := movabletype.ConvertToArticleStructs(articleMaps)

	fmt.Printf("処理対象記事数: %d\n", len(articles))

	// 記事ごとに処理
	for _, article := range articles {
		if err := h.processArticle(article, outputBaseDir); err != nil {
			return err
		}
	}

	return nil
}

// 単一記事の処理
func (h *HugoConverter) processArticle(article movabletype.Article, outputBaseDir string) error {
	// 日付情報の取得
	if article.Date == "" {
		fmt.Println("警告: DATEフィールドがない記事をスキップします")
		return nil
	}

	// 日付文字列をパース - utilパッケージを使用
	t, err := util.ParseArticleDate(article.Date)
	if err != nil {
		fmt.Printf("警告: 日付パースエラー '%s': %v - この記事をスキップします\n", article.Date, err)
		return nil
	}

	// 出力ディレクトリパスを生成 - utilパッケージを使用
	dirName := util.FormatDirName(t)
	dirPath := filepath.Join(outputBaseDir, dirName)

	// ディレクトリを作成
	if err := h.fs.MkdirAll(dirPath); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %v", err)
	}

	// Hugoデータの準備
	hugoData := h.prepareHugoData(article, t)

	// 本文の処理
	if article.Body != "" {
		markdown, err := h.converter.ConvertHTMLToMarkdown(article.Body)
		if err != nil {
			fmt.Println("警告: HTML→Markdown変換エラー:", err)
			hugoData.Content = article.Body
		} else {
			hugoData.Content = markdown
		}
	}

	// テンプレートを使って出力内容を生成
	var output strings.Builder
	if err := h.tmpl.Execute(&output, hugoData); err != nil {
		return fmt.Errorf("テンプレート実行エラー: %v", err)
	}

	// ファイルに書き込み
	filePath := filepath.Join(dirPath, "index.md")
	return h.fs.WriteFile(filePath, output.String())
}

// HugoArticleデータを準備する
func (h *HugoConverter) prepareHugoData(article movabletype.Article, t time.Time) HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	hugoData := HugoArticle{
		Title:    strings.ReplaceAll(title, "\"", "\\\""),
		Date:     t.Format("2006-01-02T15:04:05-07:00"),
		Slug:     util.CreateSlug(title), // utilパッケージを使用
		Category: article.Category,
		Image:    article.Image,
		Summary:  strings.ReplaceAll(article.Excerpt, "\"", "\\\""),
	}

	// タグの処理
	if article.Keywords != "" {
		for _, tag := range strings.Split(article.Keywords, ",") {
			hugoData.Tags = append(hugoData.Tags, strings.TrimSpace(tag))
		}
	}

	return hugoData
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
	tmpl, err := template.ParseFiles("templates/hugo.tmpl")
	if err != nil {
		// テンプレートファイルがない場合は組み込みテンプレートを使用
		fmt.Println("警告: テンプレートファイルが見つかりません。組み込みテンプレートを使用します:", err)

		// 組み込みのテンプレート定義
		const builtinTemplate = `---
title: "{{ .Title }}"
date: {{ .Date }}
slug: {{ .Slug }}
{{ if .Category }}category:
  - {{ .Category }}
{{ end }}
{{ if .Tags }}tags:
{{ range .Tags }}  - {{ . }}
{{ end }}{{ end }}
{{ if .Image }}cover:
    image: "{{ .Image }}"
    alt: "{{ .Title }}"
    hidden: true
    caption: "{{ .Title }}"
{{ end }}
draft: false
showtoc: false
{{ if .Summary }}summary: "{{ .Summary }}"{{ end }}
---

{{ .Content }}
`
		tmpl, err = template.New("hugo").Parse(builtinTemplate)
		if err != nil {
			fmt.Printf("組み込みテンプレート解析エラー: %v\n", err)
			return
		}
	}

	// コンバーターの生成
	hugoConverter := NewHugoConverter(fs, converter, tmpl)

	// 変換の実行
	filePath := os.Args[1]
	if err := hugoConverter.Convert(filePath, "output"); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		fmt.Println("変換が完了しました")
	}
}
