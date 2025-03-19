package converter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLToMarkdownConverter_ConvertHTMLToMarkdown(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "基本的なHTML変換",
			html:     "<h1>タイトル</h1><p>テキスト</p>",
			expected: "# タイトル\n\nテキスト\n",
		},
		{
			name:     "リンク変換",
			html:     "<a href=\"https://example.com\">リンク</a>",
			expected: "[リンク](https://example.com)\n",
		},
		{
			name:     "キーワードクラス付きリンク",
			html:     "<a class=\"keyword\" href=\"https://example.com\">キーワード</a>",
			expected: "キーワード\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertHTMLToMarkdown(tt.html)
			require.NoError(t, err, "変換エラー")

			// 改行の正規化（OSによる違いを吸収）
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")

			assert.Equal(t, expected, result, "変換結果が期待値と一致しません")
		})
	}
}

func TestHTMLToMarkdownConverter_FormatHTMLWithIndentation(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	tests := []struct {
		name     string
		html     string
		contains []string // 完全一致ではなく、含まれるべき文字列
	}{
		{
			name:     "基本的なHTML整形",
			html:     "<div><p>テキスト</p><ul><li>項目1</li></ul></div>",
			contains: []string{"<div>", "<p>", "テキスト", "</p>", "<ul>", "<li>", "項目1", "</li>", "</ul>", "</div>"},
		},
		{
			name:     "brタグを含むHTML",
			html:     "テキスト1<br>テキスト2<br />テキスト3",
			contains: []string{"テキスト1", "<br />", "テキスト2", "<br />", "テキスト3"},
		},
		{
			name:     "インライン属性",
			html:     "<div class=\"test\" id=\"example\">テキスト</div>",
			contains: []string{"<div", "class=\"test\"", "id=\"example\"", ">", "テキスト", "</div>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.FormatHTMLWithIndentation(tt.html)
			require.NoError(t, err, "整形エラー")

			// 結果が空でないことを確認
			assert.NotEmpty(t, result, "整形結果が空です")

			// 期待する文字列が含まれていることを確認
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected,
					"期待する文字列が結果に含まれていません: %q", expected)
			}
		})
	}
}

func TestCleanupHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "brタグの処理",
			input:    "テスト1<br>テスト2<br/>テスト3<br />テスト4",
			expected: "テスト1<br />テスト2<br />テスト3<br />テスト4",
		},
		{
			name:     "空白行の削除",
			input:    "テスト1\n   \nテスト2\n\n\nテスト3",
			expected: "テスト1\nテスト2\nテスト3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanupHTML(tt.input)

			// 改行の正規化
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")

			assert.Equal(t, expected, result, "cleanupHTML処理結果が期待値と一致しません")
		})
	}
}

func TestFormatConsecutiveEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "連続する改行の削除",
			input:    "テスト1\n\n\n\nテスト2\n\n\nテスト3",
			expected: "テスト1\n\nテスト2\n\nテスト3",
		},
		{
			name:     "前後の空白削除",
			input:    "\n\nテスト\n\n",
			expected: "テスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatConsecutiveEmptyLines(tt.input)

			// 改行の正規化
			result = strings.ReplaceAll(result, "\r\n", "\n")
			expected := strings.ReplaceAll(tt.expected, "\r\n", "\n")

			assert.Equal(t, expected, result, "空行処理結果が期待値と一致しません")
		})
	}
}

// インデント整形機能のテスト
func TestHTMLFormatting_Complex(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	// 複雑なHTMLの例
	complexHTML := `
    <div class="container">
    <header><h1>タイトル</h1><nav><ul><li>メニュー1</li><li>メニュー2</li></ul></nav></header>
    <main>
    <article>
    <h2>見出し</h2>
    <p>テキスト<br>改行後テキスト</p>
    <section class="highlight" id="section1">
    <h3>サブ見出し</h3>
    <ul><li>項目1</li><li>項目2</li></ul>
    </section>
    </article>
    </main>
    </div>
    `

	// 複数の観点からサブテストを実施
	t.Run("フォーマットが成功すること", func(t *testing.T) {
		result, err := conv.FormatHTMLWithIndentation(complexHTML)
		require.NoError(t, err, "複雑なHTMLの整形でエラーが発生しました")
		assert.NotEmpty(t, result, "複雑なHTMLの整形結果が空です")
	})

	t.Run("タグが正しくインデントされること", func(t *testing.T) {
		result, _ := conv.FormatHTMLWithIndentation(complexHTML)

		// 期待されるインデントの構造をチェック
		expectedParts := []string{
			"<div class=\"container\">",
			"  <header>",
			"    <h1>",
			"      タイトル",
			"    </h1>",
			"    <nav>",
			"      <ul>",
			"        <li>",
			"          メニュー1",
			"        </li>",
		}

		for _, part := range expectedParts {
			assert.Contains(t, result, part,
				"整形結果に期待されるインデント構造が含まれていません: %s", part)
		}
	})

	t.Run("ネストされたタグが正しくインデントされること", func(t *testing.T) {
		result, _ := conv.FormatHTMLWithIndentation(complexHTML)

		// より深くネストされた要素のインデントを確認
		deepNestedParts := []string{
			"    <article>",
			"      <h2>",
			"        見出し",
			"      </h2>",
			"      <p>",
			"        テキスト",
			"        <br />",
			"        改行後テキスト",
			"      </p>",
		}

		for _, part := range deepNestedParts {
			assert.Contains(t, result, part,
				"ネストされた要素のインデントが正しくありません: %s", part)
		}
	})

	t.Run("属性が維持されること", func(t *testing.T) {
		result, _ := conv.FormatHTMLWithIndentation(complexHTML)

		// クラスとIDが正しく維持されているか確認
		assert.Contains(t, result, "<div class=\"container\">",
			"クラス属性が維持されていません")
		assert.Contains(t, result, "<section class=\"highlight\" id=\"section1\">",
			"複数の属性が正しく維持されていません")
	})

	t.Run("brタグが正規化されること", func(t *testing.T) {
		result, _ := conv.FormatHTMLWithIndentation(complexHTML)

		// brタグが正しく<br />の形式に変換されているか確認
		assert.Contains(t, result, "<br />",
			"brタグが正規化されていません")
		assert.NotContains(t, result, "<br>",
			"brタグが正しく置換されていません")
	})
}

// 追加: 極端なケースの処理テスト
func TestHTMLFormatting_EdgeCases(t *testing.T) {
	conv := NewHTMLToMarkdownConverter()

	t.Run("空のHTML", func(t *testing.T) {
		result, err := conv.FormatHTMLWithIndentation("")
		require.NoError(t, err, "空のHTMLでエラーが発生しました")
		assert.Empty(t, result, "空のHTMLの整形結果が空でありません")
	})

	t.Run("不正なHTML", func(t *testing.T) {
		invalidHTML := "<div><p>閉じタグ忘れ</div>"
		result, err := conv.FormatHTMLWithIndentation(invalidHTML)
		require.NoError(t, err, "不正なHTMLでもエラーにならないこと")
		assert.NotEmpty(t, result, "不正なHTMLでも何らかの結果が返されること")
		// gohtml は閉じタグ忘れなどの不正なHTMLでもベストエフォートで処理
	})

	t.Run("HTMLタグのない純粋なテキスト", func(t *testing.T) {
		plainText := "これはHTMLタグを含まないただのテキストです。\n複数行あります。"
		result, err := conv.FormatHTMLWithIndentation(plainText)
		require.NoError(t, err, "プレーンテキストでエラーが発生しました")
		assert.Equal(t, plainText, result, "HTMLタグがないテキストは変更されないこと")
	})

	t.Run("極端に長い属性値", func(t *testing.T) {
		longAttrHTML := "<div class=\"" + strings.Repeat("a", 1000) + "\">テキスト</div>"
		result, err := conv.FormatHTMLWithIndentation(longAttrHTML)
		require.NoError(t, err, "長い属性値でエラーが発生しました")
		assert.Contains(t, result, "class=\""+strings.Repeat("a", 1000), "長い属性値が維持されていること")
		assert.Contains(t, result, "</div>", "タグの構造が維持されていること")
	})
}
