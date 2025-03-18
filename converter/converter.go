package converter

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
)

// HTMLConverter はHTMLをMarkdownに変換するインターフェース
type HTMLConverter interface {
	ConvertHTMLToMarkdown(html string) (string, error)
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
