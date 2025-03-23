package reporter

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Reporter はコマンドライン出力のインターフェース
type Reporter interface {
	// 警告メッセージを出力
	PrintWarning(format string, args ...interface{})
	// 進捗情報を開始
	Start(message string)
	// 進捗情報を更新
	UpdateProgress(current int, message string)
	// 進捗情報を完了
	Finish(message string)
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

// PrintWarning は警告メッセージを出力する
func (r *ProgressReporter) PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(r.writer, "\n警告: "+format+"\n", args...)
}
