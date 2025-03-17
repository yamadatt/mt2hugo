package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
)

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

	for _, article := range articles {
		// 日付情報の取得
		dateStr, ok := article["DATE"]
		if !ok {
			// DATEフィールドがない記事はスキップ（エラーを出さずに続行）
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

		writer := bufio.NewWriter(file)

		// YAML形式のフロントマターを作成
		fmt.Fprintf(writer, "---\n")

		// タイトルを出力
		title := article["TITLE"]
		if title == "" {
			title = "無題"
		}
		fmt.Fprintf(writer, "title: \"%s\"\n", strings.ReplaceAll(title, "\"", "\\\""))

		// 日付を適切な形式で出力
		fmt.Fprintf(writer, "date: %s\n", t.Format("2006-01-02T15:04:05-07:00"))

		// スラグ - タイトルから生成
		slug := strings.ReplaceAll(title, " ", "-")
		slug = strings.ReplaceAll(slug, "/", "-")
		slug = strings.ReplaceAll(slug, "\\", "-")
		slug = strings.ReplaceAll(slug, ":", "-")
		fmt.Fprintf(writer, "slug: %s\n", slug)

		// カテゴリ
		if category, ok := article["CATEGORY"]; ok && category != "" {
			fmt.Fprintf(writer, "category:\n  - %s\n", category)
		}

		// タグ（MovableTypeのキーワードがあれば）
		if keywords, ok := article["KEYWORDS"]; ok && keywords != "" {
			fmt.Fprintf(writer, "tags:\n")
			for _, tag := range strings.Split(keywords, ",") {
				fmt.Fprintf(writer, "  - %s\n", strings.TrimSpace(tag))
			}
		}

		// 画像
		if image, ok := article["IMAGE"]; ok && image != "" {
			fmt.Fprintf(writer, "cover:\n")
			fmt.Fprintf(writer, "    image: \"%s\"\n", image)
			fmt.Fprintf(writer, "    alt: \"%s\"\n", title)
			fmt.Fprintf(writer, "    hidden: true\n")
			fmt.Fprintf(writer, "    caption: \"%s\"\n", title)
		}

		// その他の設定
		fmt.Fprintf(writer, "draft: false\n")
		fmt.Fprintf(writer, "showtoc: false\n")

		// サマリーがあれば
		if excerpt, ok := article["EXCERPT"]; ok && excerpt != "" {
			fmt.Fprintf(writer, "summary: \"%s\"\n", strings.ReplaceAll(excerpt, "\"", "\\\""))
		}

		fmt.Fprintf(writer, "---\n\n")

		// BODYがあればHTMLからMarkdownに変換して追加
		if body, ok := article["BODY"]; ok {
			markdown, err := converter.ConvertString(body)
			if err != nil {
				// 変換エラー時は元のHTMLをそのまま使用
				fmt.Println("警告: HTML→Markdown変換エラー:", err)
				fmt.Fprintf(writer, "%s\n", body)
			} else {
				fmt.Fprintf(writer, "%s\n", markdown)
			}
		}

		writer.Flush()
		file.Close() // 各ファイル処理後にすぐ閉じる
	}

	return nil
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
