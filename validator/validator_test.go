package validator

import (
	"fmt"
	"mt2hugo/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テスト用のモックレポーター
type MockReporter struct {
	warnings []string
}

func (m *MockReporter) PrintWarning(format string, args ...interface{}) {
	m.warnings = append(m.warnings, fmt.Sprintf(format, args...))
}

// 以下のメソッドは reporter.Reporter インターフェースに必要
func (m *MockReporter) Start(message string)                       {}
func (m *MockReporter) UpdateProgress(current int, message string) {}
func (m *MockReporter) Finish(message string)                      {}

// 残りのコードは変更不要...

func TestArticleValidator_ValidateArticle(t *testing.T) {
	t.Run("基本的なバリデーションテスト", func(t *testing.T) {
		// テストケース定義
		testCases := []struct {
			name          string
			article       models.MTArticle
			expectError   bool
			errorContains string
		}{
			{
				name: "有効な記事",
				article: models.MTArticle{
					Title: "テスト記事",
					Date:  "2023-01-01 12:00:00",
				},
				expectError: false,
			},
			{
				name: "タイトルなし記事",
				article: models.MTArticle{
					Date: "2023-01-01 12:00:00",
				},
				expectError:   true,
				errorContains: "タイトルが空",
			},
			{
				name: "日付なし記事",
				article: models.MTArticle{
					Title: "日付なし記事",
				},
				expectError:   true,
				errorContains: "日付が空",
			},
			{
				name:        "空の記事",
				article:     models.MTArticle{},
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// モックレポーター初期化
				mockReporter := &MockReporter{}

				// バリデーター作成
				validator := NewArticleValidator(mockReporter)

				// テスト実行
				err := validator.ValidateArticle(tc.article)

				// アサーション
				if tc.expectError {
					assert.Error(t, err, "エラーが発生すべき場合にエラーがありません")
					if tc.errorContains != "" {
						assert.Contains(t, err.Error(), tc.errorContains, "エラーメッセージが期待した文字列を含んでいません")
					}
				} else {
					assert.NoError(t, err, "エラーが発生すべきでない場合にエラーが発生しました")
				}
			})
		}
	})
}

func TestArticleValidator_ValidateCategory(t *testing.T) {
	t.Run("カテゴリバリデーション", func(t *testing.T) {
		testCases := []struct {
			name            string
			category        string
			expectWarning   bool
			warningContains string
		}{
			{
				name:          "通常のカテゴリ",
				category:      "テストカテゴリ",
				expectWarning: false,
			},
			{
				name:            "特殊文字を含むカテゴリ",
				category:        "特殊@カテゴリ",
				expectWarning:   true,
				warningContains: "特殊文字",
			},
			{
				name:            "複数の特殊文字を含むカテゴリ",
				category:        "特殊@#$%カテゴリ",
				expectWarning:   true,
				warningContains: "特殊文字",
			},
			{
				name:          "空のカテゴリ",
				category:      "",
				expectWarning: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// モックレポーター初期化
				mockReporter := &MockReporter{}

				// バリデーター作成
				validator := NewArticleValidator(mockReporter)

				// カテゴリの検証実行
				validator.validateCategory(tc.category)

				// 警告のアサーション
				if tc.expectWarning {
					assert.NotEmpty(t, mockReporter.warnings, "警告が発生すべき場合に警告がありません")
					if tc.warningContains != "" && len(mockReporter.warnings) > 0 {
						assert.Contains(t, mockReporter.warnings[0], tc.warningContains,
							"警告メッセージが期待した文字列を含んでいません")
					}
				} else {
					assert.Empty(t, mockReporter.warnings, "警告が発生すべきでない場合に警告が発生しました")
				}
			})
		}
	})
}

func TestRequiredFieldRule(t *testing.T) {
	t.Run("必須フィールドルール", func(t *testing.T) {
		// テスト用の記事
		article := models.MTArticle{
			Title: "テスト記事",
			Date:  "2023-01-01",
		}

		// タイトルフィールドのルール
		titleRule := NewRequiredFieldRule("TITLE", func(a models.MTArticle) string {
			return a.Title
		})

		// 日付フィールドのルール
		dateRule := NewRequiredFieldRule("DATE", func(a models.MTArticle) string {
			return a.Date
		})

		// カテゴリフィールドのルール (必須ではない)
		categoryRule := NewRequiredFieldRule("CATEGORY", func(a models.MTArticle) string {
			return a.Category
		})

		// テスト実行と検証
		t.Run("有効なフィールド", func(t *testing.T) {
			err := titleRule.Validate(article)
			assert.NoError(t, err, "タイトルフィールドは値を持っているためエラーが発生すべきではありません")

			err = dateRule.Validate(article)
			assert.NoError(t, err, "日付フィールドは値を持っているためエラーが発生すべきではありません")
		})

		t.Run("無効なフィールド", func(t *testing.T) {
			err := categoryRule.Validate(article)
			assert.Error(t, err, "カテゴリフィールドは値を持っていないためエラーが発生すべきです")
			assert.Contains(t, err.Error(), "CATEGORYフィールドは必須です",
				"エラーメッセージが期待通りではありません")
		})
	})
}

func TestEnhancedArticleValidator(t *testing.T) {
	t.Run("拡張バリデーター機能", func(t *testing.T) {
		// モックレポーター初期化
		mockReporter := &MockReporter{}

		// 拡張バリデーター作成
		enhancedValidator := NewEnhancedArticleValidator(mockReporter)

		// 追加のルールを作成
		titleRule := NewRequiredFieldRule("TITLE", func(a models.MTArticle) string {
			return a.Title
		})

		// ルールを追加
		enhancedValidator.AddRule(titleRule)

		t.Run("すべてのルールが成功するケース", func(t *testing.T) {
			article := models.MTArticle{
				Title: "テスト記事",
				Date:  "2023-01-01",
			}

			err := enhancedValidator.ValidateArticle(article)
			assert.NoError(t, err, "すべてのフィールドが有効な場合、エラーは発生すべきではありません")
		})

		t.Run("タイトルが欠けているケース", func(t *testing.T) {
			article := models.MTArticle{
				Date: "2023-01-01",
			}

			err := enhancedValidator.ValidateArticle(article)
			assert.Error(t, err, "タイトルが欠けている場合、エラーが発生すべきです")
		})

		t.Run("日付が欠けているケース", func(t *testing.T) {
			article := models.MTArticle{
				Title: "テスト記事",
			}

			err := enhancedValidator.ValidateArticle(article)
			assert.Error(t, err, "日付が欠けている場合、エラーが発生すべきです")
		})
	})
}
