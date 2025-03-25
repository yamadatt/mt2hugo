package reporter

import (
	"bytes"
	"io"
	"os"
)

// CaptureOutput は f の実行中に os.Stdout へ出力された内容をキャプチャして返します
func CaptureOutput(f func()) string {
	// 元の標準出力を保存
	oldStdout := os.Stdout

	// パイプを作成
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 関数を実行
	f()

	// パイプを閉じて標準出力を元に戻す
	w.Close()
	os.Stdout = oldStdout

	// パイプからデータを読み取る
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
