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
func createHugoFiles(articles []map[string]string) error {
	// HTMLをMarkdownに変換するコンバーターを初期化
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

	// テンプレートを解析 - ファイルから読み込み
	tmpl, err := template.ParseFiles("templates/hugo.tmpl")
	if err != nil {
		return fmt.Errorf("テンプレート解析エラー: %v", err)
	}

	for _, article := range articles {
		// 日付情報の取得
		dateStr, ok := article["DATE"]
		if !ok {
			fmt.Println("警告: DATEフィールドがない記事をスキップします")
			continue
		}

		// 日付文字列をパース
		t, err := parseArticleDate(dateStr)
		if err != nil {
			fmt.Printf("警告: 日付パースエラー '%s': %v - この記事をスキップします\n", dateStr, err)
			continue
		}

		// ディレクトリ名を年/月/日/時分 形式で作成
		dirName := formatDirName(t)
		dirPath := filepath.Join("output", dirName)

		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}

		filePath := filepath.Join(dirPath, "index.md")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// データの準備
		hugoData := HugoArticle{
			Title:    strings.ReplaceAll(getOrDefault(article, "TITLE", "無題"), "\"", "\\\""),
			Date:     t.Format("2006-01-02T15:04:05-07:00"),
			Slug:     createSlug(getOrDefault(article, "TITLE", "無題")),
			Category: article["CATEGORY"],
			Image:    article["IMAGE"],
			Summary:  strings.ReplaceAll(article["EXCERPT"], "\"", "\\\""),
		}

		// タグの処理
		if keywords, ok := article["KEYWORDS"]; ok && keywords != "" {
			for _, tag := range strings.Split(keywords, ",") {
				hugoData.Tags = append(hugoData.Tags, strings.TrimSpace(tag))
			}
		}

		// 本文の処理
		if body, ok := article["BODY"]; ok {
			markdown, err := converter.ConvertString(body)
			if err != nil {
				fmt.Println("警告: HTML→Markdown変換エラー:", err)
				hugoData.Content = body
			} else {
				hugoData.Content = markdown
			}
		}

		// テンプレートを実行してファイルに書き込む
		if err := tmpl.Execute(file, hugoData); err != nil {
			return fmt.Errorf("テンプレート実行エラー: %v", err)
		}

		file.Close() // 各ファイル処理後にすぐ閉じる
	}

	return nil
}

// マップからキーの値を取得するヘルパー関数。存在しない場合はデフォルト値を返す
func getOrDefault(m map[string]string, key, defaultValue string) string {
	if value, ok := m[key]; ok && value != "" {
		return value
	}
	return defaultValue
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

	articles := ParseMovableTypeExportFile(lines)
	fmt.Printf("処理対象記事数: %d\n", len(articles))
	if err := createHugoFiles(articles); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		fmt.Println("変換が完了しました")
	}
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
