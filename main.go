package main

import (
	"fmt"
	"os"
	"time"

	"mt2hugo/internal/config"
	"mt2hugo/internal/errors"
	"mt2hugo/internal/factory"
	"mt2hugo/models"
	"mt2hugo/reporter"
)

func main() {
	// 設定の読み込み
	cfg, err := config.ParseFlags()
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.ErrConfiguration, "設定の解析に失敗しました")
		fmt.Fprintln(os.Stderr, wrappedErr)
		fmt.Fprintln(os.Stderr, config.GetUsage())
		os.Exit(1)
	}

	// fmt.Printf("デバッグ: NoMarkdown=%v\n", cfg.NoMarkdown)

	// 設定の警告を表示
	config.PrintFlagWarnings(cfg)

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
		if errors.Is(err, errors.ErrConfiguration) {
			fmt.Fprintln(os.Stderr, "設定エラー:", err)
			os.Exit(2)
		} else {
			fmt.Fprintln(os.Stderr, "初期化に失敗しました:", err)
			os.Exit(1)
		}
	}

	// MovableTypeのパース処理
	mtArticles, err := factory.LoadMTArticles(cfg.InputFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "処理を中止します。")
		os.Exit(1)
	}

	// 処理用コンポーネントのセットアップ
	components.SetupProcessingComponents(len(mtArticles), cfg)
	components.ProgressReporter.Start("記事変換を開始")

	// 処理前のメモリ使用量
	reporter.ReportMemoryUsage(components.ProgressReporter, "処理開始時")

	// 処理結果の追跡
	result := reporter.NewProcessResult(len(mtArticles))

	// 記事ごとに処理
	for i, article := range mtArticles {
		// 進捗更新
		components.ProgressReporter.UpdateProgress(i+1, fmt.Sprintf("記事 %d/%d を処理中", i+1, result.TotalArticles))

		// 変換処理
		err := processArticle(article, components)
		if err != nil {
			// エラータイプに基づいた処理
			if errors.Is(err, errors.ErrValidation) {
				result.AddValidationError(i+1, err)
			} else if errors.Is(err, errors.ErrTransform) {
				result.AddTransformError(i+1, err)
			} else {
				// その他のエラー
				result.AddOtherError(i+1, err)
			}
			continue
		}

		result.IncrementProcessed()
	}

	// 処理後のメモリ使用量
	reporter.ReportMemoryUsage(components.ProgressReporter, "処理終了後")

	// 処理完了
	elapsed := time.Since(start)
	components.ProgressReporter.Finish(fmt.Sprintf("処理が完了しました。所要時間: %s", elapsed))

	// 結果の出力
	reporter.ReportProcessResult(components.ProgressReporter, result)
}

func processArticle(
	article models.MTArticle,
	components *factory.Components,
) error {
	hugoArticle, dateTime, err := components.Transformer.Transform(article)
	if err != nil {
		// すでにラップされている可能性があるのでWrapIfNotTypedを使用
		return errors.WrapIfNotTyped(err, errors.ErrProcessing, "記事の変換に失敗しました")
	}

	_, err = components.FileGenerator.GenerateFile(hugoArticle, dateTime)
	if err != nil {
		return errors.WrapIfNotTyped(err, errors.ErrFileSystem, "ファイル生成に失敗しました")
	}
	return nil
}
