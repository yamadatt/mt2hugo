package validator

import (
	"fmt"
	"mt2hugo/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テスト用のモックレポーター
type MockReporter struct {
	WarningMessages []string
	InfoMessages    []string // 新しいフィールド
	Messages        []string
}

func (m *MockReporter) PrintWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.WarningMessages = append(m.WarningMessages, message)
}

// 新しいメソッド
func (m *MockReporter) PrintInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.InfoMessages = append(m.InfoMessages, message)
}

func (m *MockReporter) PrintMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.Messages = append(m.Messages, message)
}

func (m *MockReporter) Start(message string)                       {}
func (m *MockReporter) UpdateProgress(current int, message string) {}
func (m *MockReporter) Finish(message string)                      {}

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
				errorContains: "TITLEフィールドは必須です", // 「タイトルが空」を修正
			},
			{
				name: "日付なし記事",
				article: models.MTArticle{
					Title: "日付なし記事",
				},
				expectError:   true,
				errorContains: "DATEフィールドは必須です", // 「日付が空」を修正
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

// カテゴリバリデーションのテスト - 特殊文字チェックが不要なので簡素化
func TestArticleValidator_ValidateCategory(t *testing.T) {
	mockReporter := &MockReporter{
		WarningMessages: []string{},
		InfoMessages:    []string{},
		Messages:        []string{},
	}

	validator := NewArticleValidator(mockReporter)

	// 任意のカテゴリのテスト - 警告は出ないはず
	validator.ValidateCategory("test-category")
	assert.Empty(t, mockReporter.WarningMessages, "カテゴリに対して警告は出ないはず")

	// 特殊文字を含むカテゴリも同様に警告は出ないはず
	mockReporter.WarningMessages = []string{} // 一応リセット
	validator.ValidateCategory("test/category")
	assert.Empty(t, mockReporter.WarningMessages, "特殊文字を含むカテゴリでも警告は出ないはず")

	// 空のカテゴリも問題ないはず
	mockReporter.WarningMessages = []string{}
	validator.ValidateCategory("")
	assert.Empty(t, mockReporter.WarningMessages, "空のカテゴリでも警告は出ないはず")
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

// 重複していたテスト関数を1つに統合
func TestEnhancedArticleValidator(t *testing.T) {
	// 基本的な機能テスト
	t.Run("基本的な機能", func(t *testing.T) {
		mockReporter := &MockReporter{
			WarningMessages: []string{},
			InfoMessages:    []string{},
			Messages:        []string{},
		}

		validator := NewEnhancedArticleValidator(mockReporter)

		// バリデータが正しく作成されたことを確認
		assert.NotNil(t, validator, "EnhancedArticleValidatorが作成されるべき")

		// 拡張バリデータもカテゴリに対して警告を出さないことを確認
		validator.ValidateCategory("test/category")
		assert.Empty(t, mockReporter.WarningMessages, "カテゴリに対して警告は出ないはず")
	})

	// 拡張機能テスト
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
