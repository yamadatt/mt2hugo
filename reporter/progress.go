package reporter

import (
	"fmt"
	"io"
	"os"
	"time"
)

// ProgressReporter は処理の進捗状況を表示するための構造体
type ProgressReporter struct {
	totalItems int
	startTime  time.Time
	writer     io.Writer
}

// NewProgressReporter は新しいProgressReporterを作成する
func NewProgressReporter(totalItems int) *ProgressReporter {
	return &ProgressReporter{
		totalItems: totalItems,
		startTime:  time.Now(),
		writer:     os.Stdout,
	}
}

// NewProgressReporterWithWriter は指定されたwriterを使用するProgressReporterを作成する
func NewProgressReporterWithWriter(totalItems int, writer io.Writer) *ProgressReporter {
	return &ProgressReporter{
		totalItems: totalItems,
		startTime:  time.Now(),
		writer:     writer,
	}
}

// ReportProgress は現在の進捗状況を表示する
func (p *ProgressReporter) ReportProgress(currentItem int) {
	if p.totalItems <= 0 {
		return
	}

	progressPercent := float64(currentItem) / float64(p.totalItems) * 100
	elapsed := time.Since(p.startTime)

	fmt.Fprintf(p.writer, "進捗: %.1f%% (%d/%d) - 経過時間: %v\r",
		progressPercent, currentItem, p.totalItems, elapsed.Round(time.Second))
}

// ReportProgressWithMessage は進捗状況と追加メッセージを表示する
func (p *ProgressReporter) ReportProgressWithMessage(currentItem int, message string) {
	if p.totalItems <= 0 {
		return
	}

	progressPercent := float64(currentItem) / float64(p.totalItems) * 100
	elapsed := time.Since(p.startTime)

	fmt.Fprintf(p.writer, "進捗: %.1f%% (%d/%d) - 経過時間: %v - %s\r",
		progressPercent, currentItem, p.totalItems, elapsed.Round(time.Second), message)
}

// Start は処理開始を記録し、初期メッセージを表示する
func (p *ProgressReporter) Start(message string) {
	p.startTime = time.Now()
	fmt.Fprintf(p.writer, "%s - 処理対象: %d項目\n", message, p.totalItems)
}

// Finish は処理終了を報告し、統計情報を表示する
func (p *ProgressReporter) Finish() time.Duration {
	elapsed := time.Since(p.startTime)
	fmt.Fprintf(p.writer, "\n処理完了 - 所要時間: %v\n", elapsed.Round(time.Second))
	return elapsed
}

// PrintMessage は新しい行にメッセージを表示する
func (p *ProgressReporter) PrintMessage(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, "\n"+format+"\n", args...)
}

// PrintWarning は警告メッセージを表示する
func (p *ProgressReporter) PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(p.writer, "\n警告: "+format+"\n", args...)
}

// Reporter はレポート機能を提供するインターフェース
type Reporter interface {
	// ReportProgress は現在の進捗状況を表示する
	ReportProgress(currentItem int)

	// PrintMessage はメッセージを表示する
	PrintMessage(format string, args ...interface{})

	// PrintWarning は警告メッセージを表示する
	PrintWarning(format string, args ...interface{})
}
