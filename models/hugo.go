package models

// HugoArticle はHugo形式の記事データモデル
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
