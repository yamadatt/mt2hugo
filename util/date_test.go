package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArticleDate(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expected    time.Time
		shouldError bool
	}{
		{
			name:        "標準的なMT形式",
			dateStr:     "01/02/2006 15:04:05",
			expected:    time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "時間が00:00:00の形式",
			dateStr:     "12/31/2023 00:00:00",
			expected:    time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "バックスラッシュで終わる日付",
			dateStr:     "01/02/2006 15:04:05\\",
			expected:    time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "前後に空白がある日付",
			dateStr:     "  01/02/2006 15:04:05  ",
			expected:    time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
			shouldError: false,
		},
		{
			name:        "無効な日付形式",
			dateStr:     "2006-01-02 15:04:05",
			expected:    time.Time{},
			shouldError: true,
		},
		{
			name:        "空の日付文字列",
			dateStr:     "",
			expected:    time.Time{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseArticleDate(tt.dateStr)

			if tt.shouldError {
				assert.Error(t, err, "無効な日付はエラーを返すべき")
			} else {
				require.NoError(t, err, "有効な日付はエラーを返すべきでない")
				assert.Equal(t, tt.expected, actual, "パースされた日付が期待値と一致すること")
			}
		})
	}
}

func TestFormatDirName(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "標準的な日時",
			input:    time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2023/12/31/2359",
		},
		{
			name:     "1桁の月と日",
			input:    time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC),
			expected: "2023/01/02/0304",
		},
		{
			name:     "年をまたぐ日時",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2024/01/01/0000",
		},
		{
			name:     "閏年2月29日",
			input:    time.Date(2024, 2, 29, 12, 34, 56, 0, time.UTC),
			expected: "2024/02/29/1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := FormatDirName(tt.input)
			assert.Equal(t, tt.expected, actual, "フォーマットされたディレクトリ名が期待値と一致すること")
		})
	}
}

func TestCreateSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "通常のタイトル",
			title: "Hello World",
		},
		{
			name:  "特殊文字を含むタイトル",
			title: "Title with: special/characters\\and spaces",
		},
		{
			name:  "日本語タイトル",
			title: "日本語のタイトル例",
		},
		{
			name:  "空のタイトル",
			title: "",
		},
		{
			name:  "非常に長いタイトル",
			title: "This is an extremely long title that should be shortened by the slug function to ensure it doesn't exceed reasonable lengths for URLs and filesystem paths",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug := CreateSlug(tt.title)

			// CreateSlug関数はMD5ハッシュの最初の8文字を返すため、
			// 長さが8文字であることを確認
			assert.Len(t, slug, 8, "スラグの長さは8文字であるべき")

			// 同じタイトルからは同じスラグが生成されることを確認（一貫性）
			assert.Equal(t, slug, CreateSlug(tt.title), "同じタイトルからは同じスラグが生成されるべき")

			// 異なるタイトルからは異なるスラグが生成されることを確認（一意性）
			// ただし、空のタイトルの場合は除く
			if tt.title != "" && tt.title != "通常のタイトル" {
				assert.NotEqual(t, slug, CreateSlug("通常のタイトル"), "異なるタイトルからは異なるスラグが生成されるべき")
			}
		})
	}
}

// 正確性検証のための手動テスト
func TestCreateSlugExactValues(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Hello World", "b10a8d"},
		{"日本語のタイトル例", "fee64d"},
		{"", "d41d8c"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			slug := CreateSlug(tt.title)
			assert.Contains(t, slug, tt.expected, "期待されるハッシュプレフィックスを含むこと")
		})
	}
}
