package models

// MTArticle はMovableType記事のデータモデル
type MTArticle struct {
	Title         string
	Date          string
	Body          string
	ExtendedBody  string
	Category      string
	Keywords      string
	Excerpt       string
	Image         string
	Author        string
	Status        string
	AllowComments bool
	Basename      string
}
