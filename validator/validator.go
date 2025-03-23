package validator

import (
	"fmt"

	"mt2hugo/models"
	"mt2hugo/reporter"
)

// ArticleValidator はMTArticleの検証を行うインターフェース
type ArticleValidator interface {
	// ValidateArticle は記事を検証し、エラーがあれば返す
	ValidateArticle(article models.MTArticle) error

	// ValidateCategory はカテゴリを検証する
	ValidateCategory(category string)

	// AddRule はバリデーションルールを追加する
	AddRule(rule ValidationRule)
}

// ValidationRule は記事の検証ルールのインターフェース
type ValidationRule interface {
	// Validate は記事を検証し、エラーがあれば返す
	Validate(article models.MTArticle) error
}

// ArticleValidatorImpl はMTArticleの検証を行う実装構造体
type ArticleValidatorImpl struct {
	reporter reporter.Reporter
	rules    []ValidationRule
}

// NewArticleValidator は新しいArticleValidatorImplを作成する
func NewArticleValidator(reporter reporter.Reporter) *ArticleValidatorImpl {
	validator := &ArticleValidatorImpl{
		reporter: reporter,
		rules:    make([]ValidationRule, 0),
	}

	// デフォルトのバリデーションルールを追加
	validator.AddRule(NewRequiredFieldRule("TITLE", func(a models.MTArticle) string {
		return a.Title
	}))
	validator.AddRule(NewRequiredFieldRule("DATE", func(a models.MTArticle) string {
		return a.Date
	}))

	return validator
}

// ValidateArticle は記事を検証し、エラーがあれば返す
func (v *ArticleValidatorImpl) ValidateArticle(article models.MTArticle) error {
	// カテゴリのバリデーション（警告のみ）
	v.ValidateCategory(article.Category)

	// すべてのルールを適用
	for _, rule := range v.rules {
		if err := rule.Validate(article); err != nil {
			return err
		}
	}

	return nil
}

// ValidateCategory はカテゴリを検証する
func (v *ArticleValidatorImpl) ValidateCategory(category string) {
	if category == "" {
		return
	}

}

// AddRule はバリデーションルールを追加する
func (v *ArticleValidatorImpl) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// インターフェース実装の確認
var _ ArticleValidator = (*ArticleValidatorImpl)(nil)

// RequiredFieldRule は必須フィールドのバリデーションルール
type RequiredFieldRule struct {
	FieldName    string
	FieldGetter  func(article models.MTArticle) string
	ErrorMessage string
}

// NewRequiredFieldRule は新しいRequiredFieldRuleを作成する
func NewRequiredFieldRule(fieldName string, getter func(article models.MTArticle) string) *RequiredFieldRule {
	return &RequiredFieldRule{
		FieldName:    fieldName,
		FieldGetter:  getter,
		ErrorMessage: fmt.Sprintf("%sフィールドは必須です", fieldName),
	}
}

// Validate は記事の必須フィールドを検証する
func (r *RequiredFieldRule) Validate(article models.MTArticle) error {
	value := r.FieldGetter(article)
	if value == "" {
		return fmt.Errorf(r.ErrorMessage)
	}
	return nil
}

// EnhancedArticleValidator は拡張されたバリデーション機能を提供する
type EnhancedArticleValidator struct {
	*ArticleValidatorImpl
}

// NewEnhancedArticleValidator は新しいEnhancedArticleValidatorを作成する
func NewEnhancedArticleValidator(reporter reporter.Reporter) *EnhancedArticleValidator {
	baseValidator := NewArticleValidator(reporter)
	return &EnhancedArticleValidator{
		ArticleValidatorImpl: baseValidator,
	}
}

// AddCustomValidations はカスタムバリデーションを追加する
func (v *EnhancedArticleValidator) AddCustomValidations() {
	// 例: 特定の条件に基づくカスタムバリデーションルール
	// v.AddRule(NewCustomRule())
}

// 必要に応じてカスタムルールを実装
// CustomRule はカスタムバリデーションルール
// type CustomRule struct{}
//
// func NewCustomRule() *CustomRule {
//     return &CustomRule{}
// }
//
// func (r *CustomRule) Validate(article models.MTArticle) error {
//     // カスタムバリデーションロジック
//     return nil
// }
