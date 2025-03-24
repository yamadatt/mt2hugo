package factory

import (
	"fmt"
	"text/template"

	"mt2hugo/converter/html"
	"mt2hugo/fs"
	"mt2hugo/generator"
	"mt2hugo/internal/config"
	"mt2hugo/models"
	"mt2hugo/parser/mtparser"
	"mt2hugo/reporter"
	"mt2hugo/templates"
	"mt2hugo/transformer/mt2hugo"
	"mt2hugo/validator"
)

// CreateFileSystem はファイルシステムを作成
func CreateFileSystem() fs.FileSystem {
	return fs.NewRealFileSystem()
}

// CreateHTMLConverter はHTML変換器を作成
func CreateHTMLConverter() *html.HTMLToMarkdownConverter {
	return html.NewHTMLToMarkdownConverter()
}

// LoadTemplate はHugoテンプレートを読み込む
func LoadTemplate(templatePath string) (*template.Template, error) {
	tmpl, err := templates.LoadHugoTemplate(templatePath)
	if err != nil {
		return nil, fmt.Errorf("テンプレート読み込みエラー: %w", err)
	}
	return tmpl, nil
}

// CreateProgressReporter は進捗レポーターを作成
func CreateProgressReporter(articleCount int) reporter.Reporter {
	return reporter.NewProgressReporter(articleCount)
}

// CreateValidator はバリデータを作成
func CreateValidator(reporter reporter.Reporter) validator.ArticleValidator {
	return validator.NewArticleValidator(reporter)
}

// CreateTransformer は変換器を作成
func CreateTransformer(cfg *config.Config, htmlConverter *html.HTMLToMarkdownConverter, reporter reporter.Reporter, validator validator.ArticleValidator) *mt2hugo.Transformer {
	return mt2hugo.NewTransformer(
		htmlConverter,
		reporter,
		validator,
		cfg.NoMarkdown,
		cfg.FormatHTML,
		cfg.ErrorMode,
	)
}

// CreateFileGenerator はファイル生成器を作成
func CreateFileGenerator(fs fs.FileSystem, tmpl *template.Template, outputDir string) *generator.FileGenerator {
	return generator.NewFileGenerator(fs, tmpl, outputDir)
}

// LoadMTArticles はMovable Typeの記事をロード
func LoadMTArticles(filePath string) ([]models.MTArticle, error) {
	articles, err := mtparser.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Movable Typeファイル解析エラー: %w", err)
	}
	return articles, nil
}
