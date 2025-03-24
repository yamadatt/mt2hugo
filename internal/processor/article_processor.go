package processor

import (
	"fmt"

	"mt2hugo/generator"
	"mt2hugo/models"
	"mt2hugo/reporter"
	"mt2hugo/transformer/mt2hugo"
)

// ProcessResult は処理結果の統計情報
type ProcessResult struct {
	TotalArticles    int
	ProcessedCount   int
	ValidationErrors int
	ProcessingErrors int
	ErrorDetails     []string
}

// ArticleProcessor は記事の処理を担当
type ArticleProcessor struct {
	reporter    reporter.Reporter
	transformer *mt2hugo.Transformer
	generator   *generator.FileGenerator
}

// NewArticleProcessor は新しいArticleProcessorを作成
func NewArticleProcessor(
	reporter reporter.Reporter,
	transformer *mt2hugo.Transformer,
	generator *generator.FileGenerator,
) *ArticleProcessor {
	return &ArticleProcessor{
		reporter:    reporter,
		transformer: transformer,
		generator:   generator,
	}
}

// ProcessArticles は記事一覧を処理
func (p *ArticleProcessor) ProcessArticles(articles []models.MTArticle) ProcessResult {
	result := ProcessResult{
		TotalArticles: len(articles),
		ErrorDetails:  make([]string, 0),
	}

	// 記事ごとに処理
	for i, article := range articles {
		// 進捗更新
		p.reporter.UpdateProgress(i+1, fmt.Sprintf("記事 %d/%d を処理中", i+1, result.TotalArticles))

		// 変換処理
		err := p.processArticle(article)
		if err != nil {
			// エラーメッセージを作成
			errMsg := p.formatErrorMessage(i+1, article, err)
			result.ErrorDetails = append(result.ErrorDetails, errMsg)
			result.ValidationErrors++
			continue
		}

		result.ProcessedCount++
	}

	return result
}

// processArticle は1つの記事を処理
func (p *ArticleProcessor) processArticle(article models.MTArticle) error {
	hugoArticle, dateTime, err := p.transformer.Transform(article)
	if err != nil {
		return err
	}

	_, err = p.generator.GenerateFile(hugoArticle, dateTime)
	return err
}

// formatErrorMessage はエラーメッセージを整形
func (p *ArticleProcessor) formatErrorMessage(index int, article models.MTArticle, err error) string {
	dateInfo := ""
	if article.Date != "" {
		dateInfo = fmt.Sprintf("DATE: %s, ", article.Date)
	} else {
		dateInfo = "DATE: 未設定, "
	}

	titleInfo := ""
	if article.Title != "" {
		titleInfo = fmt.Sprintf("Title: %s, ", article.Title)
	} else {
		titleInfo = "Title: 未設定, "
	}

	return fmt.Sprintf("記事[%d] 変換エラー: %s%s%v", index, dateInfo, titleInfo, err)
}
