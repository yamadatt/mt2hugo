package reporter

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Reporter はレポート出力のためのインターフェース
type Reporter interface {
	// PrintMessage はメッセージを出力する
	PrintMessage(format string, args ...interface{})

	// PrintWarning は警告メッセージを出力する
	PrintWarning(format string, args ...interface{})

	// PrintInfo は情報メッセージを出力する
	PrintInfo(format string, args ...interface{})

	// Start は処理の開始を通知する
	Start(message string)

	// UpdateProgress は進捗を更新する
	UpdateProgress(current int, message string)

	// Finish は処理の終了を通知する
	Finish(message string)

	// DisplayResult はProcessResultを表示する
	DisplayResult(result *ProcessResult)
}

// ProgressReporter は進捗表示を行う構造体
type ProgressReporter struct {
	total      int
	startTime  time.Time
	lastUpdate time.Time
	writer     io.Writer
}

// NewProgressReporter は新しいProgressReporterを作成する
func NewProgressReporter(total int) *ProgressReporter {
	return &ProgressReporter{
		total:  total,
		writer: os.Stdout,
	}
}

// NewProgressReporterWithWriter は指定されたwriterを使用するProgressReporterを作成する
func NewProgressReporterWithWriter(total int, writer io.Writer) *ProgressReporter {
	return &ProgressReporter{
		total:  total,
		writer: writer,
	}
}

// Start は進捗表示を開始する
func (r *ProgressReporter) Start(message string) {
	r.startTime = time.Now()
	r.lastUpdate = r.startTime
	fmt.Fprintf(r.writer, "%s...\n", message)
}

// UpdateProgress は進捗状況を更新する
func (r *ProgressReporter) UpdateProgress(current int, message string) {
	now := time.Now()
	// 更新の間隔を制限（例: 最低0.5秒間隔）
	if now.Sub(r.lastUpdate) < 500*time.Millisecond && current < r.total {
		return
	}
	r.lastUpdate = now

	percent := float64(current) / float64(r.total) * 100
	fmt.Fprintf(r.writer, "\r%s... %d/%d (%.1f%%)   ", message, current, r.total, percent)
}

// Finish は進捗表示を完了する
func (r *ProgressReporter) Finish(message string) {
	fmt.Fprintf(r.writer, "\n%s\n", message)
}

// PrintMessage はメッセージを出力する
func (r *ProgressReporter) PrintMessage(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// PrintWarning は警告メッセージを出力する
func (r *ProgressReporter) PrintWarning(format string, args ...interface{}) {
	fmt.Printf("警告: "+format+"\n", args...)
}

// PrintInfo は情報メッセージを出力する
func (r *ProgressReporter) PrintInfo(format string, args ...interface{}) {
	fmt.Printf("情報: "+format+"\n", args...)
}

// DisplayResult はProcessResultを表示する
func (r *ProgressReporter) DisplayResult(result *ProcessResult) {
	DisplayResults(result) // 既存の関数を呼び出す
}

// ConsoleReporter はコンソールに出力するReporterの実装
type ConsoleReporter struct {
	// フィールド
}

// PrintInfo は情報メッセージを出力する
func (r *ConsoleReporter) PrintInfo(format string, args ...interface{}) {
	fmt.Printf("情報: "+format+"\n", args...)
}

// PrintMessage はメッセージを出力する
func (r *ConsoleReporter) PrintMessage(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// DisplayResult はProcessResultを表示する
func (r *ConsoleReporter) DisplayResult(result *ProcessResult) {
	DisplayResults(result) // 既存の関数を呼び出す
}
