package hugo

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"mt2hugo/converter"
	"mt2hugo/fs"
	"mt2hugo/movabletype"
	"mt2hugo/reporter" // reporter パッケージをインポート
	"mt2hugo/util"
	"mt2hugo/validator" // バリデーターパッケージをインポート
)

// HugoConverter は記事変換を行う構造体
type HugoConverter struct {
	fs         fs.FileSystem
	converter  converter.HTMLConverter
	tmpl       *template.Template
	noMarkdown bool
	formatHTML bool
	reporter   reporter.Reporter           // インターフェースに変更
	validator  *validator.ArticleValidator // バリデーターを追加
}

// NewConverter はHugoConverterの新しいインスタンスを作成する
func NewConverter(fileSystem fs.FileSystem, htmlConverter converter.HTMLConverter,
	tmpl *template.Template, reporter reporter.Reporter,
	noMarkdown bool, formatHTML bool) *HugoConverter {

	// バリデーターを作成
	articleValidator := validator.NewArticleValidator(reporter)

	return &HugoConverter{
		fs:         fileSystem,
		converter:  htmlConverter,
		tmpl:       tmpl,
		noMarkdown: noMarkdown,
		formatHTML: formatHTML,
		reporter:   reporter,
		validator:  articleValidator,
	}
}

// HugoArticle は、Hugoの記事データ構造体
type HugoArticle struct {
	Title        string
	Date         string
	Slug         string
	Category     string
	Tags         []string
	Image        string
	Summary      string
	Body         string
	ExtendedBody string
}

// ConversionReport は変換処理の結果を保持する構造体
type ConversionReport struct {
	TotalArticles    int      // 合計記事数
	ProcessedCount   int      // 正常に処理された記事数
	ValidationErrors int      // バリデーションエラー数
	ProcessingErrors int      // 処理中エラー数
	ErrorDetails     []string // エラー詳細情報
}

// Error はerrorインターフェースを実装
func (r *ConversionReport) Error() string {
	if r.ValidationErrors > 0 || r.ProcessingErrors > 0 {
		return fmt.Sprintf("変換完了: 合計 %d 記事中 %d 記事が正常に処理されました "+
			"(%d 記事がバリデーションエラー, %d 記事が処理中エラー)",
			r.TotalArticles, r.ProcessedCount, r.ValidationErrors, r.ProcessingErrors)
	}
	return fmt.Sprintf("変換完了: 合計 %d 記事がすべて正常に処理されました", r.TotalArticles)
}

// HasErrors はエラーが発生したかどうかを返す
func (r *ConversionReport) HasErrors() bool {
	return r.ValidationErrors > 0 || r.ProcessingErrors > 0
}

// レポートヘルパー関数 - 警告メッセージの出力を統一
func (h *HugoConverter) reportWarning(format string, args ...interface{}) {
	if h.reporter != nil {
		h.reporter.PrintWarning(format, args...)
	} else {
		fmt.Printf("\n警告: "+format+"\n", args...)
	}
}

// ConvertEntries は既にパース済みの記事配列をHugo形式に変換する
func (h *HugoConverter) ConvertEntries(articles []movabletype.Article, outputBaseDir string) error {
	report := &ConversionReport{
		TotalArticles: len(articles),
		ErrorDetails:  make([]string, 0),
	}

	// 記事ごとに処理（進捗表示は外部から行う）
	for i, article := range articles {
		// 記事のバリデーション
		if err := h.validator.ValidateArticle(article); err != nil {
			errMsg := fmt.Sprintf("記事[%d] バリデーションエラー: %v", i+1, err)
			h.reportWarning(errMsg)
			report.ValidationErrors++
			report.ErrorDetails = append(report.ErrorDetails, errMsg)
			continue // バリデーションに失敗した記事はスキップ
		}

		// 記事の処理
		if err := h.processArticle(article, outputBaseDir); err != nil {
			errMsg := fmt.Sprintf("記事[%d] 処理エラー: %v", i+1, err)
			h.reportWarning(errMsg)
			report.ProcessingErrors++
			report.ErrorDetails = append(report.ErrorDetails, errMsg)
		} else {
			report.ProcessedCount++
		}
	}

	// エラーが一つでもあればレポートを返す
	if report.HasErrors() {
		return report
	}

	return nil
}

// 単一記事の処理
func (h *HugoConverter) processArticle(article movabletype.Article, outputBaseDir string) error {
	// 日付文字列をパース
	t, err := util.ParseArticleDate(article.Date)
	if err != nil {
		return fmt.Errorf("日付パースエラー '%s': %v", article.Date, err)
	}

	// 出力ディレクトリパスを生成
	dirPath, err := h.createOutputDirectory(outputBaseDir, t)
	if err != nil {
		return err
	}

	// Hugoデータの準備
	hugoData, err := h.prepareArticleData(article, t)
	if err != nil {
		return err
	}

	// テンプレートを使って出力内容を生成し、ファイルに書き込み
	return h.renderAndSaveArticle(hugoData, dirPath)
}

// 出力ディレクトリを作成するヘルパーメソッド
func (h *HugoConverter) createOutputDirectory(baseDir string, date time.Time) (string, error) {
	dirName := util.FormatDirName(date)
	dirPath := filepath.Join(baseDir, dirName)

	// ディレクトリを作成
	if err := h.fs.MkdirAll(dirPath); err != nil {
		return "", fmt.Errorf("ディレクトリ作成エラー: %v", err)
	}

	return dirPath, nil
}

// 記事データを準備するヘルパーメソッド
func (h *HugoConverter) prepareArticleData(article movabletype.Article, t time.Time) (HugoArticle, error) {
	// 基本データの準備
	hugoData := createBaseHugoArticle(article, t)

	// タグの処理
	hugoData.Tags = parseTags(article.Keywords)

	// 本文の処理
	processedBody, err := h.processContent(article.Body)
	if err != nil {
		return HugoArticle{}, fmt.Errorf("本文処理エラー: %v", err)
	}
	hugoData.Body = processedBody

	// 拡張本文の処理
	if article.ExtendedBody != "" {
		processedExtendedBody, err := h.processContent(article.ExtendedBody)
		if err != nil {
			return HugoArticle{}, fmt.Errorf("拡張本文処理エラー: %v", err)
		}
		hugoData.ExtendedBody = processedExtendedBody
	}

	return hugoData, nil
}

// 基本記事データを作成する
func createBaseHugoArticle(article movabletype.Article, t time.Time) HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	// slugの決定
	slug := determineSlug(article.Basename, title)

	return HugoArticle{
		Title:        strings.ReplaceAll(title, "\"", "\\\""),
		Date:         t.Format("2006-01-02T15:04:05-07:00"),
		Slug:         slug,
		Category:     article.Category,
		Image:        article.Image,
		Summary:      strings.ReplaceAll(article.Excerpt, "\"", "\\\""),
		ExtendedBody: "", // 初期値は空文字、後で設定
	}
}

// スラグを決定する
func determineSlug(basename, title string) string {
	if basename != "" {
		return basename
	}
	return util.CreateSlug(title)
}

// タグ文字列をパースする
func parseTags(keywords string) []string {
	if keywords == "" {
		return nil
	}

	var tags []string
	for _, tag := range strings.Split(keywords, ",") {
		tags = append(tags, strings.TrimSpace(tag))
	}
	return tags
}

// テンプレートレンダリングとファイル保存
func (h *HugoConverter) renderAndSaveArticle(hugoData HugoArticle, dirPath string) error {
	// テンプレートを使って出力内容を生成
	content, err := h.renderTemplate(hugoData)
	if err != nil {
		return err
	}

	// ファイルに書き込み
	filePath := filepath.Join(dirPath, "index.md")
	return h.fs.WriteFile(filePath, content)
}

// テンプレートレンダリング
func (h *HugoConverter) renderTemplate(data HugoArticle) (string, error) {
	var output strings.Builder
	if err := h.tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("テンプレート実行エラー: %v", err)
	}
	return output.String(), nil
}

// 本文コンテンツを処理するヘルパーメソッド
func (h *HugoConverter) processContent(content string) (string, error) {
	// HTMLをそのまま出力する場合
	if h.noMarkdown {
		// 整形せずそのまま出力
		if !h.formatHTML {
			return content, nil
		}

		// HTML整形が必要な場合
		processed, err := h.converter.FormatHTMLWithIndentation(content)
		if err != nil {
			return "", fmt.Errorf("HTMLの整形エラー: %v", err)
		}
		return processed, nil
	}

	// HTMLをMarkdownに変換する場合
	processed, err := h.converter.ConvertHTMLToMarkdown(content)
	if err != nil {
		return "", fmt.Errorf("Markdown変換エラー: %v", err)
	}
	return processed, nil
}

// ProcessArticle は単一記事を処理する（公開メソッド版）
func (h *HugoConverter) ProcessArticle(article movabletype.Article, outputBaseDir string) error {
	return h.processArticle(article, outputBaseDir)
}

// HugoArticleデータを準備する
func prepareHugoData(article movabletype.Article, t time.Time) HugoArticle {
	title := article.Title
	if title == "" {
		title = "無題"
	}

	// slugの決定: BASENAMEがあれば使用し、なければタイトルからスラグを生成
	slug := ""
	if article.Basename != "" {
		slug = article.Basename
	} else {
		slug = util.CreateSlug(title)
	}

	hugoData := HugoArticle{
		Title:        strings.ReplaceAll(title, "\"", "\\\""),
		Date:         t.Format("2006-01-02T15:04:05-07:00"),
		Slug:         slug,
		Category:     article.Category,
		Image:        article.Image,
		Summary:      strings.ReplaceAll(article.Excerpt, "\"", "\\\""),
		ExtendedBody: "", // 初期値は空文字、後で設定
	}

	// タグの処理
	if article.Keywords != "" {
		for _, tag := range strings.Split(article.Keywords, ",") {
			hugoData.Tags = append(hugoData.Tags, strings.TrimSpace(tag))
		}
	}

	return hugoData
}
