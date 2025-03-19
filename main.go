package main

import (
	"fmt"
	"os"
	"time"

	"mt2hugo/converter" // converterパッケージをインポート
	"mt2hugo/fs"
	"mt2hugo/hugo"
)

// エントリーポイント
func main() {
	if len(os.Args) < 2 {
		fmt.Println("使い方: go run main.go <Movable_Typeエクスポートファイルのパス>")
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

	// コンバーターの生成
	hugoConverter := hugo.NewConverter(fileSystem, htmlConverter, tmpl)

	// 変換の実行
	filePath := os.Args[1]
	fmt.Printf("処理を開始します: %s\n", filePath)

	start := time.Now()
	if err := hugoConverter.Convert(filePath, "output"); err != nil {
		fmt.Println("Hugoファイル作成エラー:", err)
	} else {
		elapsed := time.Since(start)
		fmt.Printf("変換が完了しました（所要時間: %v）\n", elapsed)
	}
}
