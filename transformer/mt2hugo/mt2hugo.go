package mt2hugo

import (
	"fmt"
	"strings"
	"time"

	"mt2hugo/converter"  // インターフェースをインポート
	"mt2hugo/downloader" // 追加
	"mt2hugo/models"
	"mt2hugo/reporter"
	"mt2hugo/transformer"
	"mt2hugo/util"
	"mt2hugo/validator"
)

// ErrorHandlingMode はエラー発生時の挙動モード
type ErrorHandlingMode int

const (
	ReturnError ErrorHandlingMode = iota
	ReturnHTML
	ReturnPartial
)

// Transformer は移行処理を行う変換器
type Transformer struct {
	htmlConverter     converter.HTMLConverter
	reporter          reporter.Reporter
	validator         validator.ArticleValidator
	noMarkdown        bool
	formatHTML        bool
	errorHandlingMode ErrorHandlingMode
	imageDownloader   *downloader.ImageDownloader
	downloadImages    bool
}

// コンパイラによるインターフェース実装チェック
var _ transformer.ArticleTransformer = (*Transformer)(nil)

// NewTransformer は新しいTransformerを作成する
func NewTransformer(
	htmlConverter converter.HTMLConverter,
	reporter reporter.Reporter,
	validator validator.ArticleValidator,
	noMarkdown bool,
	formatHTML bool,
	errorMode ErrorHandlingMode,
	imageDownloader *downloader.ImageDownloader,
	downloadImages bool,
) *Transformer {
	return &Transformer{
		htmlConverter:     htmlConverter,
		reporter:          reporter,
		validator:         validator,
		noMarkdown:        noMarkdown,
		formatHTML:        formatHTML,
		errorHandlingMode: errorMode,
		imageDownloader:   imageDownloader,
		downloadImages:    downloadImages,
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
	processedExtendedBody := ""
	if mtArticle.ExtendedBody != "" {
		processedExtendedBody, err = t.processContent(mtArticle.ExtendedBody)
		if err != nil {
			return models.HugoArticle{}, time.Time{}, fmt.Errorf("拡張本文処理エラー: %w", err)
		}
		hugoArticle.ExtendedBody = processedExtendedBody
	}

	// IMAGEフィールドの処理
	if mtArticle.Image == "" && t.imageDownloader != nil {
		t.reporter.PrintInfo("画像ダウンローダーが利用可能です。本文から画像を検索します。")
		// 本文から画像URLを抽出
		imgURLs := t.imageDownloader.ExtractImageURLs(processedBody)
		t.reporter.PrintInfo("本文から %d 件の画像URLを抽出しました", len(imgURLs))

		// 拡張本文がある場合はそこからも抽出
		if len(imgURLs) == 0 && processedExtendedBody != "" {
			imgURLs = t.imageDownloader.ExtractImageURLs(processedExtendedBody)
			t.reporter.PrintInfo("拡張本文から %d 件の画像URLを抽出しました", len(imgURLs))
		}

		// 最初の画像を使用
		if len(imgURLs) > 0 {
			hugoArticle.Image = imgURLs[0]
			t.reporter.PrintInfo("記事「%s」の最初の画像を自動検出: %s", mtArticle.Title, imgURLs[0])
		} else {
			t.reporter.PrintInfo("記事「%s」から画像を検出できませんでした", mtArticle.Title)
		}
	} else if mtArticle.Image == "" {
		t.reporter.PrintInfo("画像ダウンローダーが利用できないため、画像の自動検出をスキップします")
	} else if mtArticle.Image != "" && t.downloadImages && t.imageDownloader != nil {
		// 画像のダウンロードはまだ行わない（出力ディレクトリが決定していないため）
		// ここではImageフィールドをそのまま設定
		hugoArticle.Image = mtArticle.Image
		t.reporter.PrintInfo("記事「%s」の既存の画像を使用: %s", mtArticle.Title, mtArticle.Image)
	} else {
		hugoArticle.Image = mtArticle.Image
	}

	return hugoArticle, dateTime, nil
}

// processContent はコンテンツを処理する（HTML→Markdown変換など）
func (t *Transformer) processContent(content string) (string, error) {
	// コンテンツが空の場合は早期リターン
	if content == "" {
		return "", nil
	}

	processedContent := content

	// Markdownへの変換処理
	if !t.noMarkdown {
		var err error
		processedContent, err = t.htmlConverter.ConvertHTMLToMarkdown(processedContent)
		if err != nil {
			switch t.errorHandlingMode {
			case ReturnError:
				return "", fmt.Errorf("Markdown変換エラー: %w", err)
			case ReturnHTML:
				t.reporter.PrintWarning("Markdown変換に失敗しました。HTMLをそのまま出力します: %v", err)
				return content, nil
			case ReturnPartial:
				t.reporter.PrintWarning("Markdown変換が一部失敗しました。部分的な結果を出力します: %v", err)
				return processedContent, nil
			}
		}
	} else if t.formatHTML {
		// HTMLフォーマット処理
		var err error
		processedContent, err = t.htmlConverter.FormatHTMLWithIndentation(processedContent)
		if err != nil {
			t.reporter.PrintWarning("HTML整形に失敗しました。整形なしで出力します: %v", err)
		}
	}

	return processedContent, nil
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
