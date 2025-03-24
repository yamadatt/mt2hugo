package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"mt2hugo/generator"
	"mt2hugo/internal/config"
	"mt2hugo/internal/errors"
	"mt2hugo/internal/factory"
	"mt2hugo/models"
	"mt2hugo/reporter"
	"mt2hugo/transformer/mt2hugo"
)

func main() {
	// 設定の読み込み
	cfg, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, config.GetUsage())
		os.Exit(1)
	}

	// 設定の警告を表示
	config.PrintFlagWarnings(cfg)

	// 初期化（ファクトリーパターン使用）
	fileSystem := factory.CreateFileSystem()
	htmlConverter := factory.CreateHTMLConverter()

	// テンプレートの読み込み
	tmpl, err := factory.LoadTemplate(cfg.TemplateFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "処理を中止します。")
		os.Exit(1)
	}

	// 変換の実行部分
	fmt.Printf("処理を開始します: %s\n", cfg.InputFile)
	if cfg.NoMarkdown {
		fmt.Println("Markdown変換を無効にしました。HTMLをそのまま出力します。")
		if cfg.FormatHTML {
			fmt.Println("HTMLを階層構造でフォーマットします。")
		}
	}

	start := time.Now()

	// コンポーネントの初期化
	components, err := factory.CreateComponents(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "初期化に失敗しました。処理を中止します。")
		os.Exit(1)
	}

	// MovableTypeのパース処理
	mtArticles, err := factory.LoadMTArticles(cfg.InputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "処理を中止します。")
		os.Exit(1)
	}

	// 処理用コンポーネントのセットアップ - 設定も渡す
	components.SetupProcessingComponents(len(mtArticles), cfg)
	components.ProgressReporter.Start("記事変換を開始")

	// 処理前のメモリ使用量
	reportMemoryUsage(components.ProgressReporter, "処理開始時")

	// バリデータの作成
	articleValidator := factory.CreateValidator(components.ProgressReporter)

	// 変換器の作成
	transformer := factory.CreateTransformer(
		cfg,
		htmlConverter,
		components.ProgressReporter,
		articleValidator,
	)

	// ファイル生成器の作成
	fileGenerator := factory.CreateFileGenerator(
		fileSystem,
		tmpl,
		cfg.OutputDir,
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
		components.ProgressReporter.UpdateProgress(i+1, fmt.Sprintf("記事 %d/%d を処理中", i+1, totalArticles))

		// 変換処理
		err := processArticle(article, transformer, fileGenerator)
		if err != nil {
			// エラータイプに基づいた処理
			if errors.Is(err, errors.ErrValidation) {
				validationErrors++
				errMsg := fmt.Sprintf("記事[%d] 検証エラー: %s", i+1, err)
				errorDetails = append(errorDetails, errMsg)
			} else if errors.Is(err, errors.ErrTransform) {
				processingErrors++
				errMsg := fmt.Sprintf("記事[%d] 変換エラー: %s", i+1, err)
				errorDetails = append(errorDetails, errMsg)
			} else {
				// その他のエラー
				errMsg := fmt.Sprintf("記事[%d] エラー: %s", i+1, err)
				errorDetails = append(errorDetails, errMsg)
			}
			continue
		}

		processedCount++
	}

	// 処理後のメモリ使用量
	reportMemoryUsage(components.ProgressReporter, "処理終了後")

	// 処理完了
	elapsed := time.Since(start)
	components.ProgressReporter.Finish(fmt.Sprintf("処理が完了しました。所要時間: %s", elapsed))

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

func reportMemoryUsage(reporter reporter.Reporter, stage string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	reporter.PrintInfo("\n%s: メモリ使用量 - ヒープ=%dMB, 合計=%dMB, システム=%dMB",
		stage,
		m.HeapAlloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024)
}
