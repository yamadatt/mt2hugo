package transformer

import (
	"time"

	"mt2hugo/models"
)

// ArticleTransformer はある形式の記事を別の形式に変換するインターフェース
type ArticleTransformer interface {
	// Transform は入力記事を変換して新しい形式の記事と日時を返す
	Transform(article interface{}) (models.HugoArticle, time.Time, error)
}
