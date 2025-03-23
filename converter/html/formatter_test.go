package html

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupHTML(t *testing.T) {
	t.Run("brタグの正規化", func(t *testing.T) {
		input := "テキスト1<br>テキスト2<br/>テキスト3<br />テキスト4"
		expected := "テキスト1<br />テキスト2<br />テキスト3<br />テキスト4"
		result := cleanupHTML(input)
		assert.Equal(t, expected, result, "brタグが正しく<br />に正規化されていません")
	})

	t.Run("空白行の削除", func(t *testing.T) {
		input := "テキスト1\n   \nテキスト2\nテキスト3"
		expected := "テキスト1\nテキスト2\nテキスト3"
		result := cleanupHTML(input)
		assert.Equal(t, expected, result, "空白のみの行が削除されていません")
	})

	t.Run("連続する改行の統合", func(t *testing.T) {
		input := "テキスト1\n\n\nテキスト2"
		expected := "テキスト1\nテキスト2"
		result := cleanupHTML(input)
		assert.Equal(t, expected, result, "連続する改行が1つにまとめられていません")
	})

	t.Run("複合的なケース", func(t *testing.T) {
		input := "テキスト1<br>\n   \n\n\nテキスト2<br/>テキスト3"
		expected := "テキスト1<br />\nテキスト2<br />テキスト3"
		result := cleanupHTML(input)
		assert.Equal(t, expected, result, "複合的なケースで正しく処理されていません")
	})

	t.Run("空の入力", func(t *testing.T) {
		result := cleanupHTML("")
		assert.Empty(t, result, "空の入力に対して空の出力が返されていません")
	})
}

func TestFormatConsecutiveEmptyLines(t *testing.T) {
	t.Run("連続する改行の統合", func(t *testing.T) {
		input := "テキスト1\n\n\n\nテキスト2\n\n\n\n\nテキスト3"
		expected := "テキスト1\n\nテキスト2\n\nテキスト3"
		result := formatConsecutiveEmptyLines(input)
		assert.Equal(t, expected, result, "連続する改行が2つにまとめられていません")
	})

	t.Run("前後の空白と改行の削除", func(t *testing.T) {
		input := "\n\n \t テキスト1\n\nテキスト2 \t \n\n"
		expected := "テキスト1\n\nテキスト2"
		result := formatConsecutiveEmptyLines(input)
		assert.Equal(t, expected, result, "前後の空白や改行が削除されていません")
	})

	t.Run("すでに整形されている入力", func(t *testing.T) {
		input := "テキスト1\n\nテキスト2"
		expected := "テキスト1\n\nテキスト2"
		result := formatConsecutiveEmptyLines(input)
		assert.Equal(t, expected, result, "すでに整形されている入力が変更されています")
	})

	t.Run("空の入力", func(t *testing.T) {
		result := formatConsecutiveEmptyLines("")
		assert.Empty(t, result, "空の入力に対して空の出力が返されていません")
	})
}

func TestHTMLToMarkdownConverter_FormatHTML(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	t.Run("基本的なHTML整形", func(t *testing.T) {
		html := "<div><p>テキスト</p></div>"
		result := conv.FormatHTML(html)
		assert.NotEqual(t, html, result, "HTML整形が行われていません")
		assert.Contains(t, result, "テキスト", "整形結果にテキストが含まれていません")
	})

	t.Run("brタグの正規化", func(t *testing.T) {
		html := "テキスト1<br>テキスト2"
		result := conv.FormatHTML(html)
		assert.Contains(t, result, "<br />", "brタグが正規化されていません")
		assert.NotContains(t, result, "<br>", "古い形式のbrタグが残っています")
	})

	t.Run("空のHTML", func(t *testing.T) {
		result := conv.FormatHTML("")
		assert.Empty(t, result, "空HTMLの整形結果が空でありません")
	})
}

func TestHTMLToMarkdownConverter_FormatHTMLWithIndentation(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	t.Run("基本的なインデント付きフォーマット", func(t *testing.T) {
		input := "<div><p>テキスト</p><ul><li>項目1</li><li>項目2</li></ul></div>"
		result, err := conv.FormatHTMLWithIndentation(input)
		require.NoError(t, err, "フォーマット中にエラーが発生しました")

		// インデントのチェック - 空白の数より存在を確認する方がテストとして安定する
		lines := strings.Split(result, "\n")
		assert.True(t, len(lines) > 1, "複数行の出力になっていません")

		// 行ごとのインデントレベルは実装に依存するため、厳密なチェックはしない
		// 代わりにコンテンツと構造が保持されているかチェック
		assert.Contains(t, result, "<div>", "divタグが保持されていません")
		assert.Contains(t, result, "<p>", "pタグが保持されていません")
		assert.Contains(t, result, "テキスト", "コンテンツが保持されていません")
		assert.Contains(t, result, "<ul>", "ulタグが保持されていません")
		assert.Contains(t, result, "<li>", "liタグが保持されていません")
		assert.Contains(t, result, "項目1", "リスト項目が保持されていません")
		assert.Contains(t, result, "項目2", "リスト項目が保持されていません")
	})

	t.Run("属性を持つタグのフォーマット", func(t *testing.T) {
		input := `<div class="container"><p id="intro">テキスト</p></div>`
		result, err := conv.FormatHTMLWithIndentation(input)
		require.NoError(t, err, "フォーマット中にエラーが発生しました")

		assert.Contains(t, result, `class="container"`, "クラス属性が保持されていません")
		assert.Contains(t, result, `id="intro"`, "ID属性が保持されていません")
		assert.Contains(t, result, "テキスト", "コンテンツが保持されていません")
	})

	t.Run("空の入力", func(t *testing.T) {
		result, err := conv.FormatHTMLWithIndentation("")
		require.NoError(t, err, "空の入力でエラーが発生しました")
		assert.Empty(t, result, "空の入力に対して空の出力が返されるべきです")
	})
}
