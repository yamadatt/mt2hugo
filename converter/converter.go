package converter

// HTMLConverter はHTMLをMarkdownに変換するインターフェース
type HTMLConverter interface {
	// ConvertHTMLToMarkdown はHTMLをMarkdownに変換する
	ConvertHTMLToMarkdown(html string) (string, error)

	// FormatHTMLWithIndentation はHTMLを整形してインデントを付ける
	FormatHTMLWithIndentation(html string) (string, error)

	// FormatHTML はHTMLを整形する
	FormatHTML(html string) string
}
