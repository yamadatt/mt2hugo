package movabletype

import (
	"fmt"
	"os"
	"strings"

	mt "github.com/yamadatt/movabletype"
)

// ParseFile はMovableTypeエクスポートファイルをパースして記事の配列を返します
// 外部パッケージ yamadatt/movabletype を使用しています
func ParseFile(filePath string) ([]Article, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイルオープンエラー: %w", err)
	}
	defer file.Close()

	// 外部パッケージを使用してパース
	entries, err := mt.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("MovableType解析エラー: %w", err)
	}

	// 外部パッケージの型から内部の型に変換
	articles := make([]Article, 0, len(entries))
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

		article := Article{
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
