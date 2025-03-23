package validator

import (
	"fmt"
	"mt2hugo/models"
	"testing"
)

// テスト用のモックレポーター
type MockReporter struct {
	warnings []string
}

func (m *MockReporter) PrintWarning(format string, args ...interface{}) {
	m.warnings = append(m.warnings, fmt.Sprintf(format, args...))
}

func (m *MockReporter) ReportProgress(current int)                      {}
func (m *MockReporter) PrintMessage(format string, args ...interface{}) {}

func TestValidateArticle(t *testing.T) {
	// テストケース
	testCases := []struct {
		name          string
		article       models.MTArticle
		expectError   bool
		expectWarning bool
	}{
		{
			name: "有効な記事",
			article: models.MTArticle{
				Title: "テスト記事",
				Date:  "2023-01-01 12:00:00",
			},
			expectError:   false,
			expectWarning: false,
		},
		{
			name: "日付なし",
			article: models.MTArticle{
				Title: "日付なし記事",
			},
			expectError:   true,
			expectWarning: false,
		},
		{
			name: "特殊文字を含むカテゴリ",
			article: models.MTArticle{
				Title:    "特殊カテゴリ記事",
				Date:     "2023-01-01 12:00:00",
				Category: "特殊@カテゴリ",
			},
			expectError:   false,
			expectWarning: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// モックレポーターの初期化
			mockReporter := &MockReporter{}

			// バリデーターの作成
			validator := NewArticleValidator(mockReporter)

			// テスト実行
			err := validator.ValidateArticle(tc.article)

			// エラーのアサーション
			if tc.expectError && err == nil {
				t.Errorf("エラーを期待していましたが、エラーがありませんでした")
			}
			if !tc.expectError && err != nil {
				t.Errorf("エラーを期待していませんでしたが、エラーがありました: %v", err)
			}

			// 警告のアサーション
			if tc.expectWarning && len(mockReporter.warnings) == 0 {
				t.Errorf("警告を期待していましたが、警告がありませんでした")
			}
			if !tc.expectWarning && len(mockReporter.warnings) > 0 {
				t.Errorf("警告を期待していませんでしたが、警告がありました: %v", mockReporter.warnings)
			}
		})
	}
}
