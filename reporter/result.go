package reporter

import (
	"fmt"
	"runtime"
)

// ProcessResult は処理結果を格納する構造体
type ProcessResult struct {
	TotalArticles    int
	ProcessedCount   int
	ValidationErrors int
	ProcessingErrors int
	ErrorDetails     []string
}

// NewProcessResult は新しいProcessResultインスタンスを作成する
func NewProcessResult(totalArticles int) *ProcessResult {
	return &ProcessResult{
		TotalArticles: totalArticles,
		ErrorDetails:  make([]string, 0),
	}
}

// AddValidationError はバリデーションエラーを追加する
func (r *ProcessResult) AddValidationError(index int, err error) {
	r.ValidationErrors++
	errMsg := fmt.Sprintf("記事[%d] 検証エラー: %s", index, err)
	r.ErrorDetails = append(r.ErrorDetails, errMsg)
}

// AddTransformError は変換エラーを追加する
func (r *ProcessResult) AddTransformError(index int, err error) {
	r.ProcessingErrors++
	errMsg := fmt.Sprintf("記事[%d] 変換エラー: %s", index, err)
	r.ErrorDetails = append(r.ErrorDetails, errMsg)
}

// AddOtherError はその他のエラーを追加する
func (r *ProcessResult) AddOtherError(index int, err error) {
	r.ProcessingErrors++
	errMsg := fmt.Sprintf("記事[%d] エラー: %s", index, err)
	r.ErrorDetails = append(r.ErrorDetails, errMsg)
}

// IncrementProcessed は正常に処理された記事数をインクリメントする
func (r *ProcessResult) IncrementProcessed() {
	r.ProcessedCount++
}

// Summary は処理結果の概要を文字列として返す
func (r *ProcessResult) Summary() string {
	return fmt.Sprintf("合計 %d 記事中 %d 記事が正常に処理されました "+
		"(%d 記事がバリデーションエラー, %d 記事が処理中エラー)",
		r.TotalArticles, r.ProcessedCount,
		r.ValidationErrors, r.ProcessingErrors)
}

// DisplayResults は処理結果を標準出力に表示する
// この関数は下位互換性のために残しておく
func DisplayResults(result *ProcessResult) {
	fmt.Printf("\n変換結果: %s\n", result.Summary())

	if len(result.ErrorDetails) > 0 {
		fmt.Println("\n以下のエラーが発生しました:")
		for _, err := range result.ErrorDetails {
			fmt.Printf("- %s\n", err)
		}
	}
}

// ReportMemoryUsage はメモリ使用状況を報告する
func ReportMemoryUsage(reporter Reporter, stage string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	reporter.PrintInfo("\n%s: メモリ使用量 - ヒープ=%dMB, 合計=%dMB, システム=%dMB",
		stage,
		m.HeapAlloc/1024/1024,
		m.TotalAlloc/1024/1024,
		m.Sys/1024/1024)
}

// ReportProcessResult は処理結果をレポートする
// Reporter.DisplayResultメソッドを使用する便利な関数
func ReportProcessResult(reporter Reporter, result *ProcessResult) {
	reporter.DisplayResult(result)
}
