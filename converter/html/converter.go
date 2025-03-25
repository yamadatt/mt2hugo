package html

import (
	"fmt"
	"mt2hugo/converter"
	"strings"

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
	// 基本的なコンバーターを作成
	converter := md.NewConverter("", true, nil)

	// カスタムルール：はてなキーワードリンクの変換
	converter.AddRules(
		md.Rule{
			Filter: []string{"a"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				// class属性を確認
				class, exists := selec.Attr("class")
				if exists && class == "keyword" {
					// キーワードリンクの場合、リンクを除去してテキストだけを返す
					return &content // 修正: ポインタを返す
				}

				// それ以外は通常の処理を行う
				href, exists := selec.Attr("href")
				if !exists {
					return &content // 修正: ポインタを返す
				}

				// 特定のドメインのリンクを処理（例: はてなブログのキーワードリンク）
				if strings.Contains(href, "d.hatena.ne.jp/keyword/") {
					// キーワードリンクと判断し、テキストのみを返す
					return &content // 修正: ポインタを返す
				}

				// 通常のリンク変換（デフォルトの処理）
				title, _ := selec.Attr("title")
				if title != "" {
					title = fmt.Sprintf(" %q", title)
				}

				// Markdown形式のリンクを返す
				result := fmt.Sprintf("[%s](%s%s)", content, href, title)
				return &result // 修正: ポインタを返す
			},
		},
	)

	// 画像変換のカスタムルールを追加
	converter.AddRules(
		md.Rule{
			Filter: []string{"img"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				// alt属性とsrc属性を取得
				alt, _ := selec.Attr("alt")
				src, exist := selec.Attr("src")
				if !exist {
					return nil
				}

				// フォーマットされる文字列からエスケープが必要な文字を処理
				alt = strings.ReplaceAll(alt, "\"", "\\\"")

				// 出力形式：シンプルな画像表示の場合はMarkdownフォーマットを維持
				if alt == "" && !selec.HasClass("hatena-asin-detail-image") &&
					!selec.HasClass("hatena-fotolife") {
					markdown := fmt.Sprintf("![%s](%s)", alt, src)
					return &markdown
				}

				// 複雑なケースはHugoショートコードを生成
				figureShortcode := fmt.Sprintf("{{< figure src=\"%s\"", src)

				if alt != "" {
					figureShortcode += fmt.Sprintf(" alt=\"%s\"", alt)
				}

				// width属性があれば追加
				if width, exists := selec.Attr("width"); exists {
					figureShortcode += fmt.Sprintf(" width=\"%s\"", width)
				}

				// height属性があれば追加
				if height, exists := selec.Attr("height"); exists {
					figureShortcode += fmt.Sprintf(" height=\"%s\"", height)
				}

				// title属性があれば追加
				if title, exists := selec.Attr("title"); exists {
					figureShortcode += fmt.Sprintf(" title=\"%s\"",
						strings.ReplaceAll(title, "\"", "\\\""))
				}

				// はてなブログ特有のクラスを検出して特別処理
				if selec.HasClass("hatena-asin-detail-image") {
					// はてな商品リンク用の特別な処理
					figureShortcode += " class=\"amazon-product\""
				} else if selec.HasClass("hatena-fotolife") {
					// はてなフォトライフ用の特別な処理
					figureShortcode += " class=\"fotolife\""
				} else if class, exists := selec.Attr("class"); exists && class != "" {
					// その他のクラス属性
					figureShortcode += fmt.Sprintf(" class=\"%s\"",
						strings.ReplaceAll(class, "\"", "\\\""))
				}

				// ショートコードを閉じる
				figureShortcode += " >}}"

				return &figureShortcode
			},
		},
	)

	// プラグインを追加
	converter.Use(plugin.GitHubFlavored())

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
