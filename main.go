package main

import (
	"flag"
	"fmt"
	"text/template"
	"time"

	"mt2hugo/converter/html" // インポートパスを更新
	"mt2hugo/fs"
	"mt2hugo/generator"
	"mt2hugo/models"
	"mt2hugo/parser/mtparser"
	"mt2hugo/reporter"
	"mt2hugo/templates"
	"mt2hugo/transformer/mt2hugo"
	"mt2hugo/validator"
)

func main() {
	// コマンドラインオプション定義
	noMarkdown := flag.Bool("no-markdown", false, "HTMLをMarkdownに変換せず、そのまま出力します")
	formatHTML := flag.Bool("format-html", false, "HTMLを階層構造でフォーマットして出力します(--no-markdownと共に使用)")
	outputDir := flag.String("output", "output", "出力先ディレクトリを指定します")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("使い方: go run main.go [オプション] <Movable_Typeエクスポートファイルのパス>")
		fmt.Println("オプション:")
		flag.PrintDefaults()
		return
	}

	// 初期化
	fileSystem := fs.NewRealFileSystem()
	htmlConverter := html.NewHTMLToMarkdownConverter() // パッケージ名を修正

	// テンプレートの読み込み
	tmpl, err := templates.LoadHugoTemplate("templates/hugo.tmpl")
	if err != nil {
		fmt.Printf("テンプレート読み込みエラー: %v\n", err)
		fmt.Println("処理を中止します。")
		return
	}

	// formatHTMLは--no-markdownと一緒に使う場合のみ効果がある
	if *formatHTML && !*noMarkdown {
		fmt.Println("注意: --format-htmlオプションは--no-markdownと一緒に使用した場合のみ効果があります")
	}

	// 変換の実行部分
	filePath := args[0]
	fmt.Printf("処理を開始します: %s\n", filePath)
	if *noMarkdown {
		fmt.Println("Markdown変換を無効にしました。HTMLをそのまま出力します。")
		if *formatHTML {
			fmt.Println("HTMLを階層構造でフォーマットします。")
		}
	}

	start := time.Now()

	// MovableTypeのパース処理
	mtArticles, err := mtparser.ParseFile(filePath)
	if err != nil {
		fmt.Printf("Movable Typeファイル解析エラー: %v\n", err)
		return
	}

	// 進捗レポーターを初期化
	progressReporter := reporter.NewProgressReporter(len(mtArticles))
	progressReporter.Start("記事変換を開始")

	// バリデータの作成
	articleValidator := validator.NewArticleValidator(progressReporter)

	// 変換器の作成
	transformer := mt2hugo.NewTransformer(
		htmlConverter,
		progressReporter,
		articleValidator,
		*noMarkdown,
		*formatHTML,
	)

	// ファイル生成器の作成
	fileGenerator := generator.NewFileGenerator(
		fileSystem,
		tmpl,
		*outputDir,
	)

	// 処理結果の追跡
	var totalArticles = len(mtArticles)
	var processedCount = 0
	var validationErrors = 0
	var processingErrors = 0
	var errorDetails = make([]string, 0)

	// 記事ごとに処理
	for i, article := range mtArticles {
		// 進捗更新
		progressReporter.UpdateProgress(i+1, fmt.Sprintf("記事 %d/%d を処理中", i+1, totalArticles))

		// 変換処理
		err := processArticle(article, transformer, fileGenerator)
		if err != nil {
			// エラーメッセージを作成し、詳細一覧に追加するだけで、ここでは表示しない
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

			errMsg := fmt.Sprintf("記事[%d] 変換エラー: %s%s%v", i+1, dateInfo, titleInfo, err)
			errorDetails = append(errorDetails, errMsg)
			validationErrors++
			continue
		}

		processedCount++
	}

	// 処理完了
	elapsed := time.Since(start)
	progressReporter.Finish(fmt.Sprintf("処理が完了しました。所要時間: %s", elapsed))

	// 結果の出力
	fmt.Printf("\n変換結果: 合計 %d 記事中 %d 記事が正常に処理されました "+
		"(%d 記事がバリデーションエラー, %d 記事が処理中エラー)\n",
		totalArticles, processedCount, validationErrors, processingErrors)

	if len(errorDetails) > 0 {
		fmt.Println("\n以下のエラーが発生しました:")
		for _, err := range errorDetails {
			fmt.Printf("- %s\n", err)
		}
	}
}

func processArticle(
	article models.MTArticle,
	transformer *mt2hugo.Transformer,
	generator *generator.FileGenerator,
) error {
	hugoArticle, dateTime, err := transformer.Transform(article)
	if err != nil {
		return err
	}

	_, err = generator.GenerateFile(hugoArticle, dateTime)
	return err
}

// mustは、エラーがあればパニックを発生させる
// templates.Mustの代わりに使用
func must(tmpl *template.Template, err error) *template.Template {
	if err != nil {
		panic(err)
	}
	return tmpl
}
