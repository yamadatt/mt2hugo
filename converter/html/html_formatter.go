package html

import (
	"regexp"
	"strings"
)

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

// isVoidElement は空要素（閉じタグが不要な要素）かどうかを判定する
func isVoidElement(tag string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[tag]
}
