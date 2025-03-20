package hugo

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
  - "{{ .Category }}"
{{ end }}
{{ if .Tags }}tags:
{{ range .Tags }}  - "{{ . }}"
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

{{ .Body }}

{{ if .ExtendedBody }}
<!--more-->

{{ .ExtendedBody }}
{{ end }}
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
	Title        string
	Date         string
	Slug         string
	Category     string
	Tags         []string
	Image        string
	Summary      string
	Body         string
	ExtendedBody string
}

// ConvertEntries は既にパース済みの記事配列をHugo形式に変換する
func (h *HugoConverter) ConvertEntries(articles []movabletype.Article, outputBaseDir string) error {
	totalArticles := len(articles)
	fmt.Printf("処理対象記事数: %d\n", totalArticles)
	startTime := time.Now()

	// 記事ごとに処理
	for i, article := range articles {
		// 現在の進捗率を計算
		progressPercent := float64(i+1) / float64(totalArticles) * 100
		elapsed := time.Since(startTime)

		// 進捗状況を表示
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

	// カテゴリに特殊文字が含まれているか確認
	invalidCharsRegex := regexp.MustCompile(`[@#$%&*!?+=/\\:;"'` + "`" + `\(\)\[\]\{\}]`)
	if invalidCharsRegex.MatchString(article.Category) {
		fmt.Printf("\n警告: カテゴリ '%s' に特殊文字が含まれています。Hugo で問題が発生する可能性があります。\n", article.Category)
	}

	// Hugoデータの準備
	hugoData := prepareHugoData(article, t)

	// 本文の処理
	var processedBody string
	if h.noMarkdown {
		// HTMLをそのまま出力（整形オプションの有無で処理が変わる）
		if h.formatHTML {
			processedBody, err = h.converter.FormatHTMLWithIndentation(article.Body)
			if err != nil {
				return fmt.Errorf("HTMLのbeautifulえらー: %v", err)
			}
		} else {
			processedBody = article.Body
		}
	} else {
		// HTMLをMarkdownに変換
		processedBody, err = h.converter.ConvertHTMLToMarkdown(article.Body)
		if err != nil {
			return fmt.Errorf("Markdown変換エラー: %v", err)
		}
	}

	// ExtendedBody の処理を追加
	var processedExtendedBody string
	if article.ExtendedBody != "" {
		if h.noMarkdown {
			// HTMLをそのまま出力（整形オプションの有無で処理が変わる）
			if h.formatHTML {
				processedExtendedBody, err = h.converter.FormatHTMLWithIndentation(article.ExtendedBody)
				if err != nil {
					return fmt.Errorf("HTMLのbeautifulえらー: %v", err)
				}
			} else {
				processedExtendedBody = article.ExtendedBody
			}
		} else {
			// HTMLをMarkdownに変換
			processedExtendedBody, err = h.converter.ConvertHTMLToMarkdown(article.ExtendedBody)
			if err != nil {
				return fmt.Errorf("Markdown変換エラー: %v", err)
			}
		}
	}

	hugoData.Body = processedBody
	hugoData.ExtendedBody = processedExtendedBody // 追加: 処理した拡張本文を設定

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
	// slug := article.Basename　//デバッグのために一旦無効化
	slug := ""
	if slug == "" {
		slug = util.CreateSlug(title)
	}

	hugoData := HugoArticle{
		Title:        strings.ReplaceAll(title, "\"", "\\\""),
		Date:         t.Format("2006-01-02T15:04:05-07:00"),
		Slug:         slug,
		Category:     article.Category,
		Image:        article.Image,
		Summary:      strings.ReplaceAll(article.Excerpt, "\"", "\\\""),
		ExtendedBody: "", // 初期値は空文字、後で設定
	}

	// タグの処理
	if article.Keywords != "" {
		for _, tag := range strings.Split(article.Keywords, ",") {
			hugoData.Tags = append(hugoData.Tags, strings.TrimSpace(tag))
		}
	}

	return hugoData
}
