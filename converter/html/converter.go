package html

import (
	"mt2hugo/converter"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
)

// HTMLToMarkdownConverter は converter.HTMLConverter インターフェースを実装
type HTMLToMarkdownConverter struct {
	converter *md.Converter
}

// 型がインターフェースを確実に実装していることを確認
var _ converter.HTMLConverter = (*HTMLToMarkdownConverter)(nil)

// NewHTMLToMarkdownConverter は新しいHTMLToMarkdownConverterを作成する
func NewHTMLToMarkdownConverter() *HTMLToMarkdownConverter {
	// デフォルトコンバーターを作成
	converter := md.NewConverter("", true, nil)

	// Github Flavored Markdownプラグインを追加
	converter.Use(plugin.GitHubFlavored())

	// カスタムルールを追加
	converter.AddRules(
		// keywordクラスを持つaタグをテキストのみに変換
		md.Rule{
			Filter: []string{"a"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				// keywordクラスを持つaタグの場合、テキストのみを返す
				if class, exists := selec.Attr("class"); exists && class == "keyword" {
					return md.String(content)
				}
				return nil
			},
		},
	)

	return &HTMLToMarkdownConverter{
		converter: converter,
	}
}

// ConvertHTMLToMarkdown はHTMLをMarkdownに変換する
func (c *HTMLToMarkdownConverter) ConvertHTMLToMarkdown(html string) (string, error) {
	if html == "" {
		return "", nil
	}

	// HTMLの事前処理
	cleanedHTML := cleanupHTML(html)

	// Markdown変換
	markdown, err := c.converter.ConvertString(cleanedHTML)
	if err != nil {
		return "", err
	}

	// 変換後のカスタム処理
	markdown = formatConsecutiveEmptyLines(markdown)

	return markdown, nil
}

// FormatHTML はHTMLを整形する
func (c *HTMLToMarkdownConverter) FormatHTML(html string) string {
	if html == "" {
		return ""
	}

	// HTMLの事前処理
	cleanedHTML := cleanupHTML(html)

	// HTMLの整形（インデントなし）
	formattedHTML := gohtml.Format(cleanedHTML)

	return formattedHTML
}

// FormatHTMLWithIndentation はHTMLを整形してインデントを付ける
func (c *HTMLToMarkdownConverter) FormatHTMLWithIndentation(html string) (string, error) {
	if html == "" {
		return "", nil
	}

	// HTMLの事前処理
	cleanedHTML := cleanupHTML(html)

	// HTMLの整形（インデント付き）
	formattedHTML := gohtml.Format(cleanedHTML)

	return formattedHTML, nil
}
