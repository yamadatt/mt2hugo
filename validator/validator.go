package validator

import (
	"fmt"
	"regexp"

	"mt2hugo/models"
	"mt2hugo/reporter"
)

// ArticleValidator は記事の検証を行う構造体
type ArticleValidator struct {
	reporter reporter.Reporter
}

// NewArticleValidator は新しいArticleValidatorを作成する
func NewArticleValidator(reporter reporter.Reporter) *ArticleValidator {
	return &ArticleValidator{
		reporter: reporter,
	}
}

// ValidateArticle は記事が有効かどうかを検証する
func (v *ArticleValidator) ValidateArticle(article models.MTArticle) error {
	// タイトルのチェック
	if article.Title == "" {
		return fmt.Errorf("タイトルが空です")
	}

	// 日付のチェック
	if article.Date == "" {
		return fmt.Errorf("日付が空です")
	}

	return nil
}

// カテゴリのバリデーション
func (v *ArticleValidator) validateCategory(category string) {
	invalidCharsRegex := regexp.MustCompile(`[@#$%&*!?+=/\\:;"'` + "`" + `\(\)\[\]\{\}]`)
	if invalidCharsRegex.MatchString(category) {
		if v.reporter != nil {
			v.reporter.PrintWarning("カテゴリ '%s' に特殊文字が含まれています。Hugo で問題が発生する可能性があります。", category)
		}
	}
}

// ValidationRule はバリデーションルールを表すインターフェース
type ValidationRule interface {
	Validate(article models.MTArticle) error
}

// RequiredFieldRule は必須フィールドのバリデーションルール
type RequiredFieldRule struct {
	FieldName    string
	FieldGetter  func(article models.MTArticle) string
	ErrorMessage string
}

// ArticleValidator の拡張版
type EnhancedArticleValidator struct {
	rules    []ValidationRule
	reporter reporter.Reporter
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

// NewEnhancedArticleValidator は新しいEnhancedArticleValidatorを作成する
func NewEnhancedArticleValidator(reporter reporter.Reporter) *EnhancedArticleValidator {
	validator := &EnhancedArticleValidator{
		reporter: reporter,
	}

	// デフォルトルールを追加
	validator.AddRule(NewRequiredFieldRule("DATE", func(a models.MTArticle) string {
		return a.Date
	}))

	return validator
}

// AddRule はバリデーションルールを追加する
func (v *EnhancedArticleValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// ValidateArticle は記事が有効かどうか検証する
func (v *EnhancedArticleValidator) ValidateArticle(article models.MTArticle) error {
	for _, rule := range v.rules {
		if err := rule.Validate(article); err != nil {
			return err
		}
	}
	return nil
}
