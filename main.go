package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	// 外部パッケージをインポート

	"mt2hugo/converter"
	"mt2hugo/fs"
	"mt2hugo/hugo"
	"mt2hugo/movabletype"
	"mt2hugo/reporter"
	"mt2hugo/templates"
)

// エントリーポイント
func main() {
	// コマンドラインオプションの定義
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
	htmlConverter := converter.NewHTMLToMarkdownConverter()

	// テンプレートの読み込み
	tmpl, err := templates.LoadHugoTemplate("templates/hugo.tmpl") // templates パッケージを使用
	if err != nil {
		fmt.Printf("テンプレート解析エラー: %v\n", err)
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
	mtArticles, err := movabletype.ParseFile(filePath)
	if err != nil {
		fmt.Printf("Movable Typeファイル解析エラー: %v\n", err)
		return
	}

	// 進捗レポーターを初期化
	progressReporter := reporter.NewProgressReporter(len(mtArticles))
	progressReporter.Start("記事変換を開始")

	// コンバーターの生成
	hugoConverter := hugo.NewConverter(
		fileSystem,
		htmlConverter,
		tmpl,
		progressReporter, // レポーターを渡す
		*noMarkdown,
		*formatHTML,
	)

	// 変換処理を実行
	err = hugoConverter.ConvertEntries(mtArticles, *outputDir)
	if err != nil {
		if report, ok := err.(*hugo.ConversionReport); ok {
			// 詳細なレポート情報を使った処理
			fmt.Printf("処理結果: %s\n", report.Error())
			if len(report.ErrorDetails) > 0 {
				fmt.Println("エラー詳細:")
				for _, detail := range report.ErrorDetails {
					fmt.Println(" - " + detail)
				}
			}
		} else {
			// 通常のエラー処理
			fmt.Printf("エラー: %v\n", err)
		}

		// エラーが発生しても異常終了しない（レポートとして処理された場合）
		if _, ok := err.(*hugo.ConversionReport); !ok {
			os.Exit(1)
		}
	} else {
		fmt.Println("変換が正常に完了しました")
	}

	// 処理完了表示はReportProgressの一部として処理される
	elapsed := time.Since(start)
	fmt.Printf("変換が完了しました（所要時間: %v）\n", elapsed)
}
