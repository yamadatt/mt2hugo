package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// 各記事用のディレクトリを作成し、Hugoファイルを書き込む
func createHugoFiles(articles []map[string]string) error {
	// HTMLをMarkdownに変換するコンバーターを初期化
	converter := md.NewConverter("", true, nil)

	for _, article := range articles {
		// 日付情報の取得
		dateStr, ok := article["DATE"]
		if !ok {
			return fmt.Errorf("missing DATE field in article")
		}

		// 日付文字列のクリーニング（余分な文字を削除）
		dateStr = strings.TrimSpace(dateStr)
		dateStr = strings.TrimRight(dateStr, "\\")

		fmt.Printf("処理中の日付文字列: %s\n", dateStr) // デバッグ用

		// 様々な日付形式を試みる
		var t time.Time
		var err error

		formats := []string{
			"01/02/2006 15:04:05", // MM/DD/YYYY HH:MM:SS
			"2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
			"01/02/06 15:04:05",   // MM/DD/YY HH:MM:SS
			"02/01/2006 15:04:05", // DD/MM/YYYY HH:MM:SS
			"2006/01/02 15:04:05", // YYYY/MM/DD HH:MM:SS
			"01/02/2006 15:04",    // MM/DD/YYYY HH:MM
		}

		for _, format := range formats {
			t, err = time.Parse(format, dateStr)
			if err == nil {
				break // 正常に解析できたらループを抜ける
			}
		}

		if err != nil {
			return fmt.Errorf("could not parse date '%s': %v", dateStr, err)
		}

		// ディレクトリ名を年/月/日/時分 形式で作成
		dirName := fmt.Sprintf("%04d/%02d/%02d/%02d%02d",
			t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())

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
		fmt.Fprintf(writer, "+++\n")
		for key, value := range article {
			// BODYはフロントマターには含めない
			if key == "BODY" {
				continue
			}
			// エスケープが必要な文字を処理
			value = strings.ReplaceAll(value, "\"", "\\\"")
			fmt.Fprintf(writer, "%s = \"%s\"\n", key, value)
		}
		fmt.Fprintf(writer, "+++\n\n")

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
