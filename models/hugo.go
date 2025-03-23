package models

// HugoArticle はHugo形式の記事データモデル
type HugoArticle struct {
	Title        string
	Date         string
	Slug         string
	Category     string   // 既存のフィールド（互換性のために残す）
	Categories   []string // 新しいフィールド - カテゴリの配列
	Tags         []string
	Image        string
	Summary      string
	Body         string
	ExtendedBody string
}
