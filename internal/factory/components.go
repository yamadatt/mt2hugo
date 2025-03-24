package factory

import (
	"fmt"
	"text/template"

	"mt2hugo/converter/html"
	"mt2hugo/fs"
	"mt2hugo/generator"
	"mt2hugo/internal/config"
	"mt2hugo/internal/errors"
	"mt2hugo/models"
	"mt2hugo/parser/mtparser"
	"mt2hugo/reporter"
	"mt2hugo/templates"
	"mt2hugo/transformer/mt2hugo"
	"mt2hugo/validator"
)

// Components はアプリケーションで使用する全てのコンポーネントを保持する構造体
type Components struct {
	FileSystem       fs.FileSystem
	HTMLConverter    *html.HTMLToMarkdownConverter
	Template         *template.Template
	ProgressReporter reporter.Reporter
	Validator        validator.ArticleValidator
	Transformer      *mt2hugo.Transformer
	FileGenerator    *generator.FileGenerator
}

// SetupProcessingComponents は記事処理に必要なコンポーネントをセットアップする
func (c *Components) SetupProcessingComponents(articleCount int, cfg *config.Config) {
	// 進捗レポーターの初期化
	c.ProgressReporter = CreateProgressReporter(articleCount)

	// バリデーターの初期化
	c.Validator = CreateValidator(c.ProgressReporter)

	// トランスフォーマーの初期化
	c.Transformer = CreateTransformer(
		cfg,
		c.HTMLConverter,
		c.ProgressReporter,
		c.Validator,
	)
}

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
		return nil, errors.Wrap(err, errors.ErrParsing,
			fmt.Sprintf("Movable Typeファイル '%s' の解析に失敗しました", filePath))
	}
	return articles, nil
}

// CreateComponents は設定に基づいて全てのコンポーネントを作成する
func CreateComponents(cfg *config.Config) (*Components, error) {
	components := &Components{}

	// ファイルシステムの作成
	components.FileSystem = CreateFileSystem()

	// HTMLコンバーターの作成
	components.HTMLConverter = CreateHTMLConverter()

	// テンプレートの読み込み
	tmpl, err := LoadTemplate(cfg.TemplateFile)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrConfiguration,
			"テンプレートの読み込みに失敗しました")
	}
	components.Template = tmpl

	// ファイル生成器の作成
	components.FileGenerator = CreateFileGenerator(
		components.FileSystem,
		components.Template,
		cfg.OutputDir,
	)

	// 進捗レポーターとバリデーターは記事数が必要なため後で初期化します
	// SetupProcessingComponentsで行います

	return components, nil
}
