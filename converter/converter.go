package converter

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
)

// HTMLConverter はHTMLをMarkdownに変換するインターフェース
type HTMLConverter interface {
	ConvertHTMLToMarkdown(html string) (string, error)
	FormatHTMLWithIndentation(html string) (string, error)
	FormatHTML(html string) string
}

// HTMLToMarkdownConverter は実際のコンバーター実装
type HTMLToMarkdownConverter struct {
	converter *md.Converter
}

// NewHTMLToMarkdownConverter は新しいHTMLToMarkdownConverterインスタンスを作成する
func NewHTMLToMarkdownConverter() *HTMLToMarkdownConverter {
	converter := md.NewConverter("", true, nil)

	// 「class="keyword"」を持つリンクのカスタムルールを追加
	converter.AddRules(md.Rule{
		Filter: []string{"a"},
		Replacement: func(content string, selec *goquery.Selection, options *md.Options) *string {
			if selec.HasClass("keyword") {
				return &content
			}
			return nil
		},
	})

	converter.Use(plugin.GitHubFlavored())

	return &HTMLToMarkdownConverter{
		converter: converter,
	}
}

// ConvertHTMLToMarkdown はHTMLをMarkdownに変換する
func (c *HTMLToMarkdownConverter) ConvertHTMLToMarkdown(html string) (string, error) {
	return c.converter.ConvertString(html)
}

// FormatHTMLWithIndentation はHTMLを整形して階層構造で出力する
// github.com/yosssi/gohtml を使用して美しくフォーマットする
func (c *HTMLToMarkdownConverter) FormatHTMLWithIndentation(htmlContent string) (string, error) {
	// まずHTMLを整形するための事前処理
	// 空白文字のみの行や不要な改行を削除
	htmlContent = cleanupHTML(htmlContent)

	// gohtmlを使用してHTMLを整形
	formatted := gohtml.Format(htmlContent)

	// 連続する空行を整える
	formatted = formatConsecutiveEmptyLines(formatted)

	return formatted, nil
}

// FormatHTML はHTMLを整形して返す
// FormatHTMLWithIndentationのラッパーとして機能し、エラーを無視します
func (c *HTMLToMarkdownConverter) FormatHTML(html string) string {
	formatted, err := c.FormatHTMLWithIndentation(html)
	if err != nil {
		// エラーが発生した場合は元のHTMLを返す
		return html
	}
	return formatted
}
