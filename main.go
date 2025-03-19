package main

import (
	"flag"
	"fmt"
	"time"

	"mt2hugo/converter"
	"mt2hugo/fs"
	"mt2hugo/hugo"
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
	tmpl, err := hugo.LoadTemplate("templates/hugo.tmpl")
	if err != nil {
		fmt.Printf("テンプレート解析エラー: %v\n", err)
		return
	}

	// formatHTMLは--no-markdownと一緒に使う場合のみ効果がある
	if *formatHTML && !*noMarkdown {
		fmt.Println("注意: --format-htmlオプションは--no-markdownと一緒に使用した場合のみ効果があります")
	}

	// コンバーターの生成
	hugoConverter := hugo.NewConverter(fileSystem, htmlConverter, tmpl, *noMarkdown, *formatHTML)

	// 変換の実行
	filePath := args[0]
	fmt.Printf("処理を開始します: %s\n", filePath)
	if *noMarkdown {
		fmt.Println("Markdown変換を無効にしました。HTMLをそのまま出力します。")
		if *formatHTML {
			fmt.Println("HTMLを階層構造でフォーマットします。")
		}
	}

	start := time.Now()
	if err := hugoConverter.Convert(filePath, *outputDir); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		elapsed := time.Since(start)
		fmt.Printf("変換が完了しました（所要時間: %v）\n", elapsed)
	}
}
