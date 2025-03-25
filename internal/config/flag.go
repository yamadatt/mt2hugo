package config

import (
	"flag"
	"fmt"

	"mt2hugo/transformer/mt2hugo"
)

// ParseFlags はコマンドライン引数を解析してConfig構造体を返す
func ParseFlags() (*Config, error) {
	cfg := NewDefaultConfig()

	// コマンドラインオプションの定義
	flag.BoolVar(&cfg.NoMarkdown, "no-markdown", cfg.NoMarkdown, "HTMLをMarkdownに変換せず、そのまま出力します")
	flag.BoolVar(&cfg.FormatHTML, "format-html", cfg.FormatHTML, "HTMLを階層構造でフォーマットして出力します(--no-markdownと共に使用)")
	flag.StringVar(&cfg.OutputDir, "output", cfg.OutputDir, "出力先ディレクトリを指定します")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "詳細なログを出力します")

	// 追加: 画像ダウンロード関連のフラグ
	flag.BoolVar(&cfg.DownloadImages, "download-images", cfg.DownloadImages, "記事内の画像をダウンロードします")
	flag.IntVar(&cfg.MaxConcurrent, "max-concurrent", cfg.MaxConcurrent, "同時ダウンロード数を指定します（デフォルト: 5）")
	flag.IntVar(&cfg.ImageTimeout, "image-timeout", cfg.ImageTimeout, "画像ダウンロードのタイムアウト秒数（デフォルト: 30）")

	// エラー処理モード
	errorModeStr := flag.String("error-mode", "html", "Markdown変換エラー時の挙動: html(デフォルト), error, partial")

	// フラグ解析の実行
	flag.Parse()

	// エラーモードの設定
	switch *errorModeStr {
	case "error":
		cfg.ErrorMode = mt2hugo.ReturnError
	case "partial":
		cfg.ErrorMode = mt2hugo.ReturnPartial
	case "html":
		cfg.ErrorMode = mt2hugo.ReturnHTML
	default:
		return nil, fmt.Errorf("不正なエラーモード: %s", *errorModeStr)
	}

	// 入力ファイルの取得
	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("入力ファイルが指定されていません\n使い方: mt2hugo [オプション] <Movable_Typeエクスポートファイルのパス>")
	}
	cfg.InputFile = args[0]

	return cfg, nil
}

// GetUsage はコマンドライン使用法の説明を返す
func GetUsage() string {
	return `使い方: mt2hugo [オプション] <Movable_Typeエクスポートファイルのパス>

オプション:
  --no-markdown      HTMLをMarkdownに変換せず、そのまま出力します
  --format-html      HTMLを階層構造でフォーマットして出力します(--no-markdownと共に使用)
  --output <DIR>     出力先ディレクトリを指定します (デフォルト: "output")
  --error-mode MODE  Markdown変換エラー時の挙動: html(デフォルト), error, partial
  --verbose          詳細なログを出力します
  --download-images  記事内の画像をダウンロードします
  --max-concurrent   同時ダウンロード数を指定します（デフォルト: 5）
  --image-timeout    画像ダウンロードのタイムアウト秒数（デフォルト: 30）
  --help             使用法を表示します
`
}

// PrintFlagWarnings は設定の警告をコンソールに出力する
func PrintFlagWarnings(cfg *Config) {
	warnings := cfg.Validate()
	for _, warning := range warnings {
		fmt.Println(warning)
	}
}
