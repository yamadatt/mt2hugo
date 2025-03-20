package validator

import (
	"fmt"
	"regexp"

	"mt2hugo/movabletype"
	"mt2hugo/reporter"
)

// ArticleValidator は記事のバリデーションを担当する
type ArticleValidator struct {
	reporter reporter.Reporter
}

// NewArticleValidator は新しいArticleValidatorを作成する
func NewArticleValidator(reporter reporter.Reporter) *ArticleValidator {
	return &ArticleValidator{
		reporter: reporter,
	}
}

// ValidationRule はバリデーションルールのインターフェース
type ValidationRule interface {
	Validate(article movabletype.Article) error
}

// RequiredFieldRule は必須フィールドを検証するルール
type RequiredFieldRule struct {
	FieldName    string
	FieldGetter  func(article movabletype.Article) string
	ErrorMessage string
}

// ValidateArticle は記事が有効かどうか検証する
func (v *ArticleValidator) ValidateArticle(article movabletype.Article) error {
	// 日付の検証
	if article.Date == "" {
		title := "無題"
		if article.Title != "" {
			title = article.Title
		}
		return fmt.Errorf("DATEフィールドがありません (記事: %s)", title)
	}

	// カテゴリに特殊文字が含まれているか確認
	v.validateCategory(article.Category)

	// TODO: 他のバリデーションルールを追加

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

// 必要に応じてその他のバリデーションメソッドを追加

// ArticleValidator の拡張版
type EnhancedArticleValidator struct {
	rules    []ValidationRule
	reporter reporter.Reporter
}

// NewRequiredFieldRule は新しいRequiredFieldRuleを作成する
func NewRequiredFieldRule(fieldName string, getter func(article movabletype.Article) string) *RequiredFieldRule {
	return &RequiredFieldRule{
		FieldName:    fieldName,
		FieldGetter:  getter,
		ErrorMessage: fmt.Sprintf("%sフィールドは必須です", fieldName),
	}
}

// Validate は記事の必須フィールドを検証する
func (r *RequiredFieldRule) Validate(article movabletype.Article) error {
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
	validator.AddRule(NewRequiredFieldRule("DATE", func(a movabletype.Article) string {
		return a.Date
	}))

	return validator
}

// AddRule はバリデーションルールを追加する
func (v *EnhancedArticleValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// ValidateArticle は記事が有効かどうか検証する
func (v *EnhancedArticleValidator) ValidateArticle(article movabletype.Article) error {
	for _, rule := range v.rules {
		if err := rule.Validate(article); err != nil {
			return err
		}
	}
	return nil
}
