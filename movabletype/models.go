package movabletype

import "strings"

// MovableTypeArticle は、Movable Typeのエクスポートデータの記事を表す構造体
type Article struct {
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
	Basename      string // BASENAMEフィールドを追加
}

// マップ形式の記事データを構造体に変換
func ConvertToArticleStructs(articleMaps []map[string]string) []Article {
	articles := make([]Article, 0, len(articleMaps))

	for _, articleMap := range articleMaps {
		article := Article{
			Title:    articleMap["TITLE"],
			Date:     articleMap["DATE"],
			Body:     articleMap["BODY"],
			Category: articleMap["CATEGORY"],
			Keywords: articleMap["KEYWORDS"],
			Excerpt:  articleMap["EXCERPT"],
			Image:    articleMap["IMAGE"],
			Author:   articleMap["AUTHOR"],
			Status:   articleMap["STATUS"],
			Basename: articleMap["BASENAME"], // BASENAMEを追加
		}

		// ALLOW_COMMENTSがある場合はbool値に変換
		if allowComments, ok := articleMap["ALLOW_COMMENTS"]; ok {
			article.AllowComments = (allowComments == "1" || strings.ToLower(allowComments) == "true")
		}

		articles = append(articles, article)
	}

	return articles
}
