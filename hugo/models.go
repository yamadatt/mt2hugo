package hugo

// HugoArticle はHugoの記事データを表す構造体
type Article struct {
	Title    string
	Date     string
	Slug     string
	Category string
	Tags     []string
	Image    string
	Summary  string
	Content  string
}
