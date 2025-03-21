package models

// HugoArticle はHugo形式の記事データを表す構造体
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
