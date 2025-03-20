package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	mt "github.com/yamadatt/movabletype" // 外部パッケージをインポート

	"mt2hugo/converter"
	"mt2hugo/fs"
	"mt2hugo/hugo"
	"mt2hugo/movabletype" // 自前のパッケージを追加
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

	// MovableTypeファイルのパース処理を行う (2つの方法)

	// 方法1: yamadatt/movabletype パッケージを使用

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	entries, err := mt.Parse(file)
	if err != nil {
		fmt.Println("Movable Typeファイル解析エラー:", err)
		return
	}

	// 解析した記事をHugo形式へ変換
	// 外部パッケージの型から自前の型に変換
	mtArticles := make([]movabletype.Article, 0, len(entries))
	for _, entry := range entries {
		// AllowCommentsの型を確認して変換
		var allowComments bool
		if entry.AllowComments == 1 {
			allowComments = true
		}

		// Date を文字列に変換
		dateStr := ""
		if !entry.Date.IsZero() {
			dateStr = entry.Date.Format("01/02/2006 15:04:05") // MT形式の日付文字列に変換
		}

		// Category をカンマ区切りの文字列に変換
		categoryStr := ""
		if len(entry.Category) > 0 {
			categoryStr = strings.Join(entry.Category, ", ")
		}

		article := movabletype.Article{
			Title:         entry.Title,
			Date:          dateStr, // 文字列に変換
			Body:          entry.Body,
			Category:      categoryStr, // 文字列に変換
			Keywords:      entry.Keywords,
			Excerpt:       entry.Excerpt,
			Image:         entry.Image,
			Author:        entry.Author,
			Status:        entry.Status,
			AllowComments: allowComments,
			Basename:      entry.Basename,
		}
		mtArticles = append(mtArticles, article)
	}

	// 変換した記事を処理
	if err := hugoConverter.ConvertEntries(mtArticles, *outputDir); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		elapsed := time.Since(start)
		fmt.Printf("変換が完了しました（所要時間: %v）\n", elapsed)
	}

	// 方法2: 自前のパーサーを使用する場合（コメントアウト）
	/*
		// ファイル読み込み
		lines, err := fileSystem.ReadFile(filePath)
		if err != nil {
			fmt.Printf("エクスポートファイル読み込みエラー: %v\n", err)
			return
		}

		// パース処理
		articleMaps, parseErrors := movabletype.ParseExportFile(lines)

		// パースエラーの処理
		if len(parseErrors) > 0 {
			fmt.Println("パース中に以下の警告が発生しました:")
			for _, parseError := range parseErrors {
				fmt.Printf("- %v\n", parseError)
			}
		}

		// 記事データを構造体に変換
		mtArticles := movabletype.ConvertToArticleStructs(articleMaps)

		// 記事を変換
		if err := hugoConverter.ConvertEntries(mtArticles, *outputDir); err != nil {
			fmt.Println("Hugoファイル作成エラー:", err)
		} else {
			elapsed := time.Since(start)
			fmt.Printf("変換が完了しました（所要時間: %v）\n", elapsed)
		}
	*/
}
