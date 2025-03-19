package converter

import (
	"bytes"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml" // 新しく追加するライブラリ
	"golang.org/x/net/html"
)

// HTMLConverter はHTMLをMarkdownに変換するインターフェース
type HTMLConverter interface {
	ConvertHTMLToMarkdown(html string) (string, error)
	FormatHTMLWithIndentation(html string) (string, error)
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

// cleanupHTML はHTMLを整形前に前処理を行う
func cleanupHTML(html string) string {
	// <br>タグを一時的に置き換え（改行コントロールのため）
	html = strings.ReplaceAll(html, "<br>", "%%%BR%%%")
	html = strings.ReplaceAll(html, "<br />", "%%%BR%%%")
	html = strings.ReplaceAll(html, "<br/>", "%%%BR%%%")

	// 空白行を削除
	whiteSpaceRegex := regexp.MustCompile(`(?m)^\s+$`)
	html = whiteSpaceRegex.ReplaceAllString(html, "")

	// 連続する改行を1つにまとめる
	multipleNewlinesRegex := regexp.MustCompile(`\n{2,}`)
	html = multipleNewlinesRegex.ReplaceAllString(html, "\n")

	// <br>タグを復元
	html = strings.ReplaceAll(html, "%%%BR%%%", "<br />")

	return html
}

// formatConsecutiveEmptyLines は連続する空行を整形する
func formatConsecutiveEmptyLines(formatted string) string {
	// 連続する空行を1つの空行にまとめる
	re := regexp.MustCompile(`\n{3,}`)
	formatted = re.ReplaceAllString(formatted, "\n\n")

	// 先頭と末尾の不要な改行を削除
	formatted = strings.TrimSpace(formatted)

	return formatted
}

// formatNodeWithIndent は再帰的にノードをインデントして整形する
func formatNodeWithIndent(n *html.Node, buf *bytes.Buffer, level int) error {
	indent := strings.Repeat("  ", level)

	switch n.Type {
	case html.DocumentNode:
		// ドキュメントノードの子ノードを処理
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if err := formatNodeWithIndent(child, buf, level); err != nil {
				return err
			}
		}
		return nil

	case html.ElementNode:
		// <html>, <head>, <body>タグは出力しない
		if n.Data == "html" || n.Data == "head" || n.Data == "body" {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				if err := formatNodeWithIndent(child, buf, level); err != nil {
					return err
				}
			}
			return nil
		}

		// 開始タグを書き込む
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(n.Data)

		// 属性を追加
		for _, attr := range n.Attr {
			buf.WriteString(" ")
			buf.WriteString(attr.Key)
			buf.WriteString("=\"")
			buf.WriteString(attr.Val)
			buf.WriteString("\"")
		}

		// 空要素タグの場合
		if isVoidElement(n.Data) {
			buf.WriteString(" />")
			buf.WriteString("\n")
			return nil
		}

		buf.WriteString(">")
		buf.WriteString("\n")

		// 子ノードがあるか確認
		hasChildElements := false
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				hasChildElements = true
				break
			}
		}

		// 子ノードを再帰的に処理
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode && strings.TrimSpace(child.Data) != "" {
				// テキストノードで内容がある場合
				textIndent := indent
				if hasChildElements {
					textIndent += "  "
				}
				buf.WriteString(textIndent)
				buf.WriteString(strings.TrimSpace(child.Data))
				buf.WriteString("\n")
			} else if child.Type != html.TextNode || strings.TrimSpace(child.Data) != "" {
				// テキストノード以外、または内容のあるテキストノード
				if err := formatNodeWithIndent(child, buf, level+1); err != nil {
					return err
				}
			}
		}

		// 閉じタグ
		buf.WriteString(indent)
		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">")
		buf.WriteString("\n")

		return nil

	case html.TextNode:
		// インラインテキストは親ノードで処理するため、ここでは何もしない
		return nil

	case html.CommentNode:
		buf.WriteString(indent)
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->")
		buf.WriteString("\n")
		return nil

	default:
		return nil
	}
}

// isVoidElement は空要素（閉じタグが不要な要素）かどうかを判定する
func isVoidElement(tag string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[tag]
}
