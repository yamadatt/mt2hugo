package html

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTMLToMarkdownConverter(t *testing.T) {
	t.Run("インスタンス作成", func(t *testing.T) {
		converter := NewHTMLToMarkdownConverter()
		assert.NotNil(t, converter, "コンバーターのインスタンスがnilです")
		assert.IsType(t, &HTMLToMarkdownConverter{}, converter, "期待する型と一致しません")
	})
}

func TestHTMLToMarkdownConverter_ConvertHTMLToMarkdown(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	t.Run("基本的なHTML変換", func(t *testing.T) {
		html := "<h1>タイトル</h1><p>テキスト</p>"
		expected := "# タイトル\n\nテキスト"

		result, err := conv.ConvertHTMLToMarkdown(html)
		require.NoError(t, err, "変換エラーが発生しました")
		assert.Equal(t, expected, result, "変換結果が期待値と一致しません")
	})

	t.Run("リンク変換", func(t *testing.T) {
		html := "<a href=\"https://example.com\">リンク</a>"
		expected := "[リンク](https://example.com)"

		result, err := conv.ConvertHTMLToMarkdown(html)
		require.NoError(t, err, "変換エラーが発生しました")
		assert.Equal(t, expected, result, "リンクの変換結果が期待値と一致しません")
	})

	t.Run("キーワードクラス付きリンク", func(t *testing.T) {
		html := "<a class=\"keyword\" href=\"https://example.com\">キーワード</a>"
		expected := "キーワード"

		result, err := conv.ConvertHTMLToMarkdown(html)
		require.NoError(t, err, "変換エラーが発生しました")
		assert.Equal(t, expected, result, "キーワードクラス付きリンクの変換結果が期待値と一致しません")
	})

	t.Run("空のHTML", func(t *testing.T) {
		result, err := conv.ConvertHTMLToMarkdown("")
		require.NoError(t, err, "空HTMLの変換でエラーが発生しました")
		assert.Empty(t, result, "空のHTMLの変換結果が空ではありません")
	})

	t.Run("テーブル変換", func(t *testing.T) {
		html := `<table><tr><th>見出し1</th><th>見出し2</th></tr><tr><td>データ1</td><td>データ2</td></tr></table>`
		result, err := conv.ConvertHTMLToMarkdown(html)
		require.NoError(t, err, "テーブル変換でエラーが発生しました")

		// Markdownでのテーブル形式をチェック
		assert.Contains(t, result, "見出し1", "テーブルの見出しが含まれていません")
		assert.Contains(t, result, "見出し2", "テーブルの見出しが含まれていません")
		assert.Contains(t, result, "データ1", "テーブルのデータが含まれていません")
		assert.Contains(t, result, "データ2", "テーブルのデータが含まれていません")
	})

	t.Run("画像変換", func(t *testing.T) {
		html := `<img src="image.jpg" alt="代替テキスト">`
		result, err := conv.ConvertHTMLToMarkdown(html)
		require.NoError(t, err, "画像変換でエラーが発生しました")
		assert.Contains(t, result, "![代替テキスト](image.jpg)", "画像変換が正しくありません")
	})
}
