package mt2hugo

import (
	"fmt"
	"strings"
	"time"

	"mt2hugo/converter" // インターフェースをインポート
	"mt2hugo/models"
	"mt2hugo/reporter"
	"mt2hugo/transformer"
	"mt2hugo/util"
	"mt2hugo/validator"
)

// ErrorHandlingMode はエラー発生時の挙動モード
type ErrorHandlingMode int

const (
	// ReturnHTML はエラー時に元のHTMLを返す
	ReturnHTML ErrorHandlingMode = iota
	// ReturnError はエラー時にエラーを返す（処理を中止する）
	ReturnError
	// ReturnPartial はエラー時に部分的な変換結果を返す
	ReturnPartial
)

// Transformer は移行処理を行う変換器
type Transformer struct {
	htmlConverter     converter.HTMLConverter // インターフェースを使用
	reporter          reporter.Reporter
	validator         validator.ArticleValidator
	noMarkdown        bool
	formatHTML        bool
	errorHandlingMode ErrorHandlingMode // 新しいフィールド
}

var _ transformer.ArticleTransformer = (*Transformer)(nil) // インターフェース実装チェック

// NewTransformer は新しいTransformerを作成する
func NewTransformer(
	htmlConverter converter.HTMLConverter, // インターフェースを使用
	reporter reporter.Reporter,
	validator validator.ArticleValidator,
	noMarkdown bool,
	formatHTML bool,
	errorMode ErrorHandlingMode, // 新しいパラメータ
) *Transformer {
	return &Transformer{
		htmlConverter:     htmlConverter,
		reporter:          reporter,
		validator:         validator,
		noMarkdown:        noMarkdown,
		formatHTML:        formatHTML,
		errorHandlingMode: errorMode,
	}
}

// SetErrorHandlingMode はエラー処理モードを設定する
func (t *Transformer) SetErrorHandlingMode(mode ErrorHandlingMode) {
	t.errorHandlingMode = mode
}

// Transform はMovableType記事をHugo記事に変換する
func (t *Transformer) Transform(article interface{}) (models.HugoArticle, time.Time, error) {
	// 型アサーションで入力を確認
	mtArticle, ok := article.(models.MTArticle)
	if !ok {
		return models.HugoArticle{}, time.Time{}, fmt.Errorf("不正な記事タイプ: %T", article)
	}

	// 記事のバリデーション
	if err := t.validator.ValidateArticle(mtArticle); err != nil {
		return models.HugoArticle{}, time.Time{}, fmt.Errorf("バリデーションエラー: %w", err)
	}

	// 日付文字列をパース
	dateTime, err := util.ParseArticleDate(mtArticle.Date)
	if err != nil {
		return models.HugoArticle{}, time.Time{}, fmt.Errorf("日付パースエラー '%s': %w", mtArticle.Date, err)
	}

	// 基本データの準備
	hugoArticle := t.createBaseHugoArticle(mtArticle, dateTime)

	// タグの処理
	hugoArticle.Tags = t.parseTags(mtArticle.Keywords)

	// 本文の処理
	processedBody, err := t.processContent(mtArticle.Body)
	if err != nil {
		return models.HugoArticle{}, time.Time{}, fmt.Errorf("本文処理エラー: %w", err)
	}
	hugoArticle.Body = processedBody

	// 拡張本文の処理
	if mtArticle.ExtendedBody != "" {
		processedExtendedBody, err := t.processContent(mtArticle.ExtendedBody)
		if err != nil {
			return models.HugoArticle{}, time.Time{}, fmt.Errorf("拡張本文処理エラー: %w", err)
		}
		hugoArticle.ExtendedBody = processedExtendedBody
	}

	return hugoArticle, dateTime, nil
}

// processContent はコンテンツを処理する（HTML→Markdown変換など）
func (t *Transformer) processContent(content string) (string, error) {
	// コンテンツが空の場合は早期リターン
	if content == "" {
		return "", nil
	}

	// Markdown変換が無効の場合
	if t.noMarkdown {
		if t.formatHTML {
			formatted, err := t.htmlConverter.FormatHTMLWithIndentation(content)
			if err != nil {
				if t.reporter != nil {
					t.reporter.PrintWarning("HTML整形エラー: %w", err)
				}
				// HTML整形に失敗した場合は元のHTMLを返す
				return content, nil
			}
			return formatted, nil
		}
		return content, nil
	}

	// Markdown変換を試みる
	converted, err := t.htmlConverter.ConvertHTMLToMarkdown(content)

	// エラーが発生した場合、エラー処理モードに応じて対応
	if err != nil {
		if t.reporter != nil {
			t.reporter.PrintWarning("Markdown変換エラー: %w", err)
		}

		switch t.errorHandlingMode {
		case ReturnError:
			// エラーを呼び出し元に返す（処理中止）
			if t.reporter != nil {
				t.reporter.PrintInfo("エラー処理モード: 処理を中止します")
			}
			return "", fmt.Errorf("Markdown変換エラー: %w", err)
		}
	}

	// 変換成功の場合
	return converted, nil
}

// createBaseHugoArticle は基本的なHugo記事データを作成する
func (t *Transformer) createBaseHugoArticle(article models.MTArticle, dateTime time.Time) models.HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	// slugの決定
	slug := t.determineSlug(article.Basename, title)

	// カテゴリを解析
	categories := t.parseCategories(article.Category)

	// 最初のカテゴリを単一カテゴリとして使用（互換性のため）
	category := ""
	if len(categories) > 0 {
		category = categories[0]
	}

	return models.HugoArticle{
		Title:        strings.ReplaceAll(title, "\"", "\\\""),
		Date:         dateTime.Format("2006-01-02T15:04:05-07:00"),
		Slug:         slug,
		Category:     category,   // 従来の単一カテゴリ（互換性のため）
		Categories:   categories, // 新しいカテゴリ配列
		Image:        article.Image,
		Summary:      strings.ReplaceAll(article.Excerpt, "\"", "\\\""),
		ExtendedBody: "", // 初期値は空文字
	}
}

// determineSlug はスラグを決定する
func (t *Transformer) determineSlug(basename, title string) string {
	if basename != "" {
		return basename
	}
	return util.CreateSlug(title)
}

// parseTags はタグ文字列をパースする
func (t *Transformer) parseTags(keywords string) []string {
	if keywords == "" {
		return nil
	}

	var tags []string
	for _, tag := range strings.Split(keywords, ",") {
		tags = append(tags, strings.TrimSpace(tag))
	}
	return tags
}

// parseCategories はカテゴリ文字列をパースする
func (t *Transformer) parseCategories(categoryStr string) []string {
	if categoryStr == "" {
		return nil
	}

	var categories []string
	for _, cat := range strings.Split(categoryStr, ",") {
		trimmed := strings.TrimSpace(cat)
		if trimmed != "" {
			categories = append(categories, trimmed)
		}
	}
	return categories
}
