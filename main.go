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
)

// MovableTypeArticle は、Movable Typeのエクスポートデータの記事を表す構造体
type MovableTypeArticle struct {
	Title         string
	Date          string
	Body          string
	Category      string
	Keywords      string
	Excerpt       string
	Image         string
	Author        string
	Status        string
	AllowComments bool
	// 他に必要なフィールドがあれば追加
}

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

// 各記事用のディレクトリを作成し、Hugoファイルを書き込む
func createHugoFiles(articles []MovableTypeArticle) error {
	// HTMLをMarkdownに変換するコンバーターを初期化
	converter := initializeConverter()

	// テンプレートを解析
	tmpl, err := loadTemplate("templates/hugo.tmpl")
	if err != nil {
		return err
	}

	for _, article := range articles {
		if err := processArticle(article, converter, tmpl); err != nil {
			return err
		}
	}

	return nil
}

// HTMLをMarkdownに変換するコンバーターを初期化
func initializeConverter() *md.Converter {
	converter := md.NewConverter("", true, nil)

	// 「class="keyword"」を持つリンクのカスタムルールを追加
	converter.AddRules(md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
			// クラス属性がkeywordのリンクをチェック
			if selec.HasClass("keyword") {
				return &content
			}
			return nil
		},
	})

	// プラグインを追加（テーブルなどの変換を改善）
	converter.Use(plugin.GitHubFlavored())

	return converter
}

// テンプレートを読み込む
func loadTemplate(templatePath string) (*template.Template, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("テンプレート解析エラー: %v", err)
	}
	return tmpl, nil
}

// 単一記事を処理する
func processArticle(article MovableTypeArticle, converter *md.Converter, tmpl *template.Template) error {
	// 日付情報の取得
	if article.Date == "" {
		fmt.Println("警告: DATEフィールドがない記事をスキップします")
		return nil
	}

	// 日付文字列をパース
	t, err := parseArticleDate(article.Date)
	if err != nil {
		fmt.Printf("警告: 日付パースエラー '%s': %v - この記事をスキップします\n", article.Date, err)
		return nil
	}

	// ファイルとディレクトリを準備
	filePath, err := prepareOutputDirectory(t)
	if err != nil {
		return err
	}

	// Hugoデータを準備
	hugoData := prepareHugoData(article, t)

	// 本文の処理
	if article.Body != "" {
		hugoData.Content = convertBody(article.Body, converter)
	}

	// ファイルに書き込む
	return writeArticleToFile(filePath, hugoData, tmpl)
}

// 出力ディレクトリを準備し、ファイルパスを返す
func prepareOutputDirectory(t time.Time) (string, error) {
	dirName := formatDirName(t)
	dirPath := filepath.Join("output", dirName)

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", err
	}

	return filepath.Join(dirPath, "index.md"), nil
}

// MovableTypeArticleからHugoArticleへ変換
func prepareHugoData(article MovableTypeArticle, t time.Time) HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	hugoData := HugoArticle{
		Title:    strings.ReplaceAll(title, "\"", "\\\""),
		Date:     t.Format("2006-01-02T15:04:05-07:00"),
		Slug:     createSlug(title),
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

// 本文をHTMLからMarkdownに変換
func convertBody(body string, converter *md.Converter) string {
	markdown, err := converter.ConvertString(body)
	if err != nil {
		fmt.Println("警告: HTML→Markdown変換エラー:", err)
		return body
	}
	return markdown
}

// ファイルに記事を書き込む
func writeArticleToFile(filePath string, hugoData HugoArticle, tmpl *template.Template) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := tmpl.Execute(file, hugoData); err != nil {
		return fmt.Errorf("テンプレート実行エラー: %v", err)
	}

	return nil
}

// タイトルからスラグを作成するヘルパー関数
func createSlug(title string) string {
	slug := strings.ReplaceAll(title, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	slug = strings.ReplaceAll(slug, ":", "-")
	return slug
}

// 上記の関数を呼び出して変換を実行するメイン関数
func main() {
	if len(os.Args) < 2 {
		fmt.Println("使い方: go run main.go <Movable_Typeエクスポートファイルのパス>")
		return
	}

	filePath := os.Args[1]
	lines, err := ReadExportFile(filePath)
	if err != nil {
		fmt.Println("エクスポートファイル読み込みエラー:", err)
		return
	}

	// map[string]stringの記事データを構造体の配列に変換
	articleMaps := ParseMovableTypeExportFile(lines)
	articles := convertToArticleStructs(articleMaps)

	fmt.Printf("処理対象記事数: %d\n", len(articles))
	if err := createHugoFiles(articles); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		fmt.Println("変換が完了しました")
	}
}

// マップ形式の記事データを構造体に変換
func convertToArticleStructs(articleMaps []map[string]string) []MovableTypeArticle {
	articles := make([]MovableTypeArticle, 0, len(articleMaps))

	for _, articleMap := range articleMaps {
		article := MovableTypeArticle{
			Title:    articleMap["TITLE"],
			Date:     articleMap["DATE"],
			Body:     articleMap["BODY"],
			Category: articleMap["CATEGORY"],
			Keywords: articleMap["KEYWORDS"],
			Excerpt:  articleMap["EXCERPT"],
			Image:    articleMap["IMAGE"],
			Author:   articleMap["AUTHOR"],
			Status:   articleMap["STATUS"],
		}

		// ALLOW_COMMENTSがある場合はbool値に変換
		if allowComments, ok := articleMap["ALLOW_COMMENTS"]; ok {
			article.AllowComments = (allowComments == "1" || strings.ToLower(allowComments) == "true")
		}

		articles = append(articles, article)
	}

	return articles
}

func parseArticleDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	dateStr = strings.TrimRight(dateStr, "\\")
	formats := []string{
		"01/02/2006 15:04:05",
		"2006-01-02 15:04:05",
		"01/02/06 15:04:05",
		"02/01/2006 15:04:05",
		"2006/01/02 15:04:05",
		"01/02/2006 15:04",
	}
	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return t, fmt.Errorf("could not parse date '%s': %v", dateStr, err)
}

func formatDirName(t time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}
