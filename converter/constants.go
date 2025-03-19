package converter

// HTMLタグと正規化に関する定数
const (
	// HTML置換用の一時的なマーカー
	brTagMarker = "%%%BR%%%"

	// 標準化されたBRタグ
	standardizedBrTag = "<br />"
)

// void要素（閉じタグが不要な要素）のリスト
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// インデントに使用する文字
const indentChar = "  "

// 無視するトップレベルタグ
var skipTopLevelTags = map[string]bool{
	"html": true, "head": true, "body": true,
}
