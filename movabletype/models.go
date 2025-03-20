package movabletype

// Article はMovableTypeの記事を表す構造体です
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
	Basename      string
}
