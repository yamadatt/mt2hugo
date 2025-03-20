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

// ConvertEntries は既にパース済みの記事配列をHugo形式に変換する
func (h *HugoConverter) ConvertEntries(articles []movabletype.Article, outputBaseDir string) error {
	// 記事ごとに処理（進捗表示は外部から行う）
	for _, article := range articles {
		// 記事のバリデーション
		if err := h.validator.ValidateArticle(article); err != nil {
			if h.reporter != nil {
				h.reporter.PrintWarning("記事バリデーションエラー: %v", err)
			} else {
				fmt.Printf("\n警告: 記事バリデーションエラー: %v\n", err)
			}
			continue // バリデーションに失敗した記事はスキップ
		}

		// 記事の処理
		if err := h.processArticle(article, outputBaseDir); err != nil {
			if h.reporter != nil {
				h.reporter.PrintWarning("記事処理エラー: %v", err)
			} else {
				fmt.Printf("\n警告: 記事処理エラー: %v\n", err)
			}
			// エラーが発生しても処理を続行
		}
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
	dirName := util.FormatDirName(t)
	dirPath := filepath.Join(outputBaseDir, dirName)

	// ディレクトリを作成
	if err := h.fs.MkdirAll(dirPath); err != nil {
		return fmt.Errorf("ディレクトリ作成エラー: %v", err)
	}

	// Hugoデータの準備
	hugoData := prepareHugoData(article, t)

	// 本文の処理
	var processedBody string
	if h.noMarkdown {
		// HTMLをそのまま出力（整形オプションの有無で処理が変わる）
		if h.formatHTML {
			processedBody, err = h.converter.FormatHTMLWithIndentation(article.Body)
			if err != nil {
				return fmt.Errorf("HTMLのbeautifulえらー: %v", err)
			}
		} else {
			processedBody = article.Body
		}
	} else {
		// HTMLをMarkdownに変換
		processedBody, err = h.converter.ConvertHTMLToMarkdown(article.Body)
		if err != nil {
			return fmt.Errorf("Markdown変換エラー: %v", err)
		}
	}

	// ExtendedBody の処理を追加
	var processedExtendedBody string
	if article.ExtendedBody != "" {
		if h.noMarkdown {
			// HTMLをそのまま出力（整形オプションの有無で処理が変わる）
			if h.formatHTML {
				processedExtendedBody, err = h.converter.FormatHTMLWithIndentation(article.ExtendedBody)
				if err != nil {
					return fmt.Errorf("HTMLのbeautifulえらー: %v", err)
				}
			} else {
				processedExtendedBody = article.ExtendedBody
			}
		} else {
			// HTMLをMarkdownに変換
			processedExtendedBody, err = h.converter.ConvertHTMLToMarkdown(article.ExtendedBody)
			if err != nil {
				return fmt.Errorf("Markdown変換エラー: %v", err)
			}
		}
	}

	hugoData.Body = processedBody
	hugoData.ExtendedBody = processedExtendedBody // 追加: 処理した拡張本文を設定

	// テンプレートを使って出力内容を生成
	var output strings.Builder
	if err := h.tmpl.Execute(&output, hugoData); err != nil {
		return fmt.Errorf("テンプレート実行エラー: %v", err)
	}

	// ファイルに書き込み
	filePath := filepath.Join(dirPath, "index.md")
	return h.fs.WriteFile(filePath, output.String())
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
	// slug := article.Basename　//デバッグのために一旦無効化
	slug := ""
	if slug == "" {
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
