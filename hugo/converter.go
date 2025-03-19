package hugo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"mt2hugo/converter" // converterパッケージをインポート
	"mt2hugo/fs"
	"mt2hugo/movabletype"
	"mt2hugo/util"
)

// HugoConverter は記事変換を行う構造体
type HugoConverter struct {
	fs         fs.FileSystem
	converter  converter.HTMLConverter // インターフェース型を使用
	tmpl       *template.Template
	noMarkdown bool // Markdownへの変換をスキップするフラグ
	formatHTML bool // HTMLを階層構造でフォーマットするフラグ
}

// デフォルトのテンプレート文字列
const DefaultTemplate = `---
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

// LoadTemplate はHugoテンプレートを読み込む
func LoadTemplate(templatePath string) (*template.Template, error) {
	// テンプレートファイルが存在するか確認
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// テンプレートファイルがない場合はデフォルトテンプレートを使用
		return template.New("hugo").Parse(DefaultTemplate)
	}

	// テンプレートファイルを読み込む
	return template.ParseFiles(templatePath)
}

// NewConverter はHugoConverterの新しいインスタンスを作成する
func NewConverter(fileSystem fs.FileSystem, htmlConverter converter.HTMLConverter, tmpl *template.Template, noMarkdown bool, formatHTML bool) *HugoConverter {
	return &HugoConverter{
		fs:         fileSystem,
		converter:  htmlConverter,
		tmpl:       tmpl,
		noMarkdown: noMarkdown,
		formatHTML: formatHTML,
	}
}

// HugoArticle は、Hugoの記事データ構造体
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

// Convert はファイルを変換するメイン処理
func (h *HugoConverter) Convert(inputPath string, outputBaseDir string) error {
	// ファイル読み込み
	lines, err := h.fs.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("エクスポートファイル読み込みエラー: %v", err)
	}

	// パース処理
	articleMaps := movabletype.ParseExportFile(lines)
	articles := movabletype.ConvertToArticleStructs(articleMaps)

	totalArticles := len(articles)
	fmt.Printf("処理対象記事数: %d\n", totalArticles)
	startTime := time.Now()

	// 記事ごとに処理
	for i, article := range articles {
		// 現在の進捗率を計算
		progressPercent := float64(i+1) / float64(totalArticles) * 100
		elapsed := time.Since(startTime)

		// 進捗状況を表示（記事タイトルなし）
		fmt.Printf("進捗: %.1f%% (%d/%d) - 経過時間: %v\r",
			progressPercent, i+1, totalArticles, elapsed.Round(time.Second))

		if err := h.processArticle(article, outputBaseDir); err != nil {
			fmt.Printf("\n警告: 記事処理エラー: %v\n", err)
			// エラーが発生しても処理を続行
		}
	}

	// 処理完了後、改行を入れて見やすくする
	fmt.Println()
	return nil
}

// 単一記事の処理
func (h *HugoConverter) processArticle(article movabletype.Article, outputBaseDir string) error {
	// 日付情報の取得
	if article.Date == "" {
		// エラーメッセージに記事のタイトルと識別情報を追加
		title := "無題"
		if article.Title != "" {
			title = article.Title
		}

		return fmt.Errorf("DATEフィールドがありません (記事: %s)", title)

	}

	// 日付文字列をパース
	t, err := util.ParseArticleDate(article.Date)
	if err != nil {
		return fmt.Errorf("日付パースエラー '%s': %v", article.Date, err)
	}

	// 出力ディレクトリパスを生成
	dirName := util.FormatDirName(t)
	dirPath := filepath.Join(outputBaseDir, dirName)

	// ディレクトリを作成
	if err := h.fs.MkdirAll(dirPath); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %v", err)
	}

	// Hugoデータの準備
	hugoData := prepareHugoData(article, t)

	// 本文の処理
	if article.Body != "" {
		if h.noMarkdown {
			// Markdown変換をスキップ
			if h.formatHTML {
				// HTMLを階層構造でフォーマット
				formattedHTML, err := h.converter.FormatHTMLWithIndentation(article.Body)
				if err != nil {
					fmt.Println("警告: HTML整形エラー:", err)
					hugoData.Content = article.Body
				} else {
					hugoData.Content = formattedHTML
				}
			} else {
				// そのままのHTMLを使用
				hugoData.Content = article.Body
			}
		} else {
			// HTMLをMarkdownに変換
			markdown, err := h.converter.ConvertHTMLToMarkdown(article.Body)
			if err != nil {
				fmt.Println("警告: HTML→Markdown変換エラー:", err)
				hugoData.Content = article.Body
			} else {
				hugoData.Content = markdown
			}
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
func prepareHugoData(article movabletype.Article, t time.Time) HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	// slugの決定: BASENAMEがあれば使用し、なければタイトルからスラグを生成
	slug := article.Basename
	if slug == "" {
		slug = util.CreateSlug(title)
	}

	hugoData := HugoArticle{
		Title:    strings.ReplaceAll(title, "\"", "\\\""),
		Date:     t.Format("2006-01-02T15:04:05-07:00"),
		Slug:     slug,
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
