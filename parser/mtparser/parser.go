package mtparser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"mt2hugo/models"

	mt "github.com/yamadatt/movabletype"
)

// ParseFile はMovableTypeエクスポートファイルをパースして記事の配列を返す
func ParseFile(filePath string) ([]models.MTArticle, error) {
	// ファイルを読み込む
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルオープンエラー: %w", err)
	}

	// 必要なフィールドの形式を修正
	content := string(fileContent)

	// ALLOW PINGSフィールドの置換 - スペースの問題を解決
	rePings := regexp.MustCompile(`(?m)^ALLOW PINGS:.*$`)
	content = rePings.ReplaceAllString(content, "ALLOW PINGS:0") // スペースなし

	// ALLOW COMMENTSフィールドも同様に修正
	reComments := regexp.MustCompile(`(?m)^ALLOW COMMENTS:.*$`)
	content = reComments.ReplaceAllString(content, "ALLOW COMMENTS:1") // スペースなし

	// デバッグ用：サニタイズ前後の変化を確認（必要に応じてコメント解除）
	/*
		if content != string(fileContent) {
			fmt.Println("ファイルの内容がサニタイズされました。")
		}
	*/

	// サニタイズのログ記録
	if content != string(fileContent) {
		fmt.Printf("情報: ファイル %s の内容がサニタイズされました\n", filePath)

		// より詳細なログ（debug用）
		/*
			originalLines := strings.Split(string(fileContent), "\n")
			modifiedLines := strings.Split(content, "\n")
			for i := 0; i < len(originalLines) && i < len(modifiedLines); i++ {
				if originalLines[i] != modifiedLines[i]) {
					fmt.Printf("  行 %d: '%s' -> '%s'\n", i+1, originalLines[i], modifiedLines[i])
				}
			}
		*/
	}

	// 修正したコンテンツで解析
	reader := strings.NewReader(content)
	entries, err := mt.Parse(reader)
	if err != nil {
		// より詳細なエラー情報を提供
		return nil, fmt.Errorf("MovableType解析エラー（ファイル: %s）: %w", filePath, err)
	}

	// 外部パッケージの型から内部モデルに変換
	articles := make([]models.MTArticle, 0, len(entries))
	for _, entry := range entries {
		// AllowCommentsの変換
		var allowComments bool
		if entry.AllowComments == 1 {
			allowComments = true
		}

		// Date を文字列に変換
		dateStr := ""
		if !entry.Date.IsZero() {
			dateStr = entry.Date.Format("01/02/2006 15:04:05") // MT形式
		}

		// Category をカンマ区切りの文字列に変換
		categoryStr := ""
		if len(entry.Category) > 0 {
			categoryStr = strings.Join(entry.Category, ", ")
		}

		article := models.MTArticle{
			Title:         entry.Title,
			Date:          dateStr,
			Body:          entry.Body,
			ExtendedBody:  entry.ExtendedBody,
			Category:      categoryStr,
			Keywords:      entry.Keywords,
			Excerpt:       entry.Excerpt,
			Image:         entry.Image,
			Author:        entry.Author,
			Status:        entry.Status,
			AllowComments: allowComments,
			Basename:      entry.Basename,
		}
		articles = append(articles, article)
	}

	return articles, nil
}
