package reporter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummyReporter は ReportMemoryUsage や ReportProcessResult のテスト用に Reporter インターフェースを実装したダミーです。
type dummyReporter struct {
	output         []string
	Info           string         // 追加: Infoメッセージをキャプチャするためのフィールド
	DisplayCalled  bool           // 追加: DisplayResultが呼ばれたかどうかをトラッキング
	ReceivedResult *ProcessResult // 追加: 受け取ったProcessResultを保存
}

// 必要なメソッドをすべて実装
func (d *dummyReporter) PrintMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	d.output = append(d.output, message)
}

func (d *dummyReporter) PrintWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	d.output = append(d.output, "警告: "+message)
}

func (d *dummyReporter) PrintInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	d.Info = message // 修正: Info文字列にメッセージを保存
	d.output = append(d.output, "情報: "+message)
}

func (d *dummyReporter) Start(message string) {
	d.output = append(d.output, "開始: "+message)
}

func (d *dummyReporter) UpdateProgress(current int, message string) {
	d.output = append(d.output, fmt.Sprintf("進捗: %d - %s", current, message))
}

func (d *dummyReporter) Finish(message string) {
	d.output = append(d.output, "完了: "+message)
}

func (d *dummyReporter) DisplayResult(result *ProcessResult) {
	d.DisplayCalled = true    // 修正: DisplayCalledをtrueに設定
	d.ReceivedResult = result // 修正: 受け取ったResultを保存
	d.output = append(d.output, "結果表示")
}

// --- ProcessResult に対するテスト ---

func TestProcessResult(t *testing.T) {
	t.Run("NewProcessResult", func(t *testing.T) {
		pr := NewProcessResult(10)
		assert.Equal(t, 10, pr.TotalArticles)
		assert.Equal(t, 0, pr.ProcessedCount)
		assert.Equal(t, 0, pr.ValidationErrors)
		assert.Equal(t, 0, pr.ProcessingErrors)
		require.NotNil(t, pr.ErrorDetails)
		assert.Len(t, pr.ErrorDetails, 0)
	})

	tests := []struct {
		name  string
		setup func(pr *ProcessResult)
		check func(t *testing.T, pr *ProcessResult)
	}{
		{
			name: "AddValidationError",
			setup: func(pr *ProcessResult) {
				pr.AddValidationError(2, fmt.Errorf("invalid title"))
			},
			check: func(t *testing.T, pr *ProcessResult) {
				assert.Equal(t, 1, pr.ValidationErrors)
				assert.Len(t, pr.ErrorDetails, 1)
				expected := "記事[2] 検証エラー: invalid title"
				assert.Equal(t, expected, pr.ErrorDetails[0])
			},
		},
		{
			name: "AddTransformError",
			setup: func(pr *ProcessResult) {
				pr.AddTransformError(3, fmt.Errorf("transform failed"))
			},
			check: func(t *testing.T, pr *ProcessResult) {
				assert.Equal(t, 1, pr.ProcessingErrors)
				assert.Len(t, pr.ErrorDetails, 1)
				expected := "記事[3] 変換エラー: transform failed"
				assert.Equal(t, expected, pr.ErrorDetails[0])
			},
		},
		{
			name: "AddOtherError",
			setup: func(pr *ProcessResult) {
				pr.AddOtherError(4, fmt.Errorf("other error"))
			},
			check: func(t *testing.T, pr *ProcessResult) {
				assert.Equal(t, 1, pr.ProcessingErrors)
				assert.Len(t, pr.ErrorDetails, 1)
				expected := "記事[4] エラー: other error"
				assert.Equal(t, expected, pr.ErrorDetails[0])
			},
		},
		{
			name: "IncrementProcessed",
			setup: func(pr *ProcessResult) {
				pr.IncrementProcessed()
				pr.IncrementProcessed()
			},
			check: func(t *testing.T, pr *ProcessResult) {
				assert.Equal(t, 2, pr.ProcessedCount)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr := NewProcessResult(10)
			tc.setup(pr)
			tc.check(t, pr)
		})
	}

	t.Run("Summary", func(t *testing.T) {
		pr := NewProcessResult(10)
		pr.IncrementProcessed()
		pr.IncrementProcessed()
		pr.AddValidationError(1, fmt.Errorf("val error"))
		pr.AddTransformError(2, fmt.Errorf("trans error"))
		summary := pr.Summary()
		expected := "合計 10 記事中 2 記事が正常に処理されました (1 記事がバリデーションエラー, 1 記事が処理中エラー)"
		assert.Equal(t, expected, summary)
	})
}

func TestDisplayResults(t *testing.T) {
	tests := []struct {
		name               string
		pr                 *ProcessResult
		expectedSubstrings []string
	}{
		{
			name: "No errors",
			pr: func() *ProcessResult {
				pr := NewProcessResult(5)
				pr.IncrementProcessed()
				return pr
			}(),
			expectedSubstrings: []string{
				"合計 5 記事中 1 記事が正常に処理されました (0 記事がバリデーションエラー, 0 記事が処理中エラー)",
			},
		},
		{
			name: "With errors",
			pr: func() *ProcessResult {
				pr := NewProcessResult(3)
				pr.IncrementProcessed()
				pr.AddValidationError(0, fmt.Errorf("invalid format"))
				pr.AddOtherError(1, fmt.Errorf("other issue"))
				return pr
			}(),
			expectedSubstrings: []string{
				"合計 3 記事中 1 記事が正常に処理されました (1 記事がバリデーションエラー, 1 記事が処理中エラー)",
				"以下のエラーが発生しました:",
				"記事[0] 検証エラー: invalid format",
				"記事[1] エラー: other issue",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := CaptureOutput(func() {
				DisplayResults(tc.pr)
			})
			for _, substr := range tc.expectedSubstrings {
				assert.Contains(t, output, substr, "出力に %q が含まれていること", substr)
			}
		})
	}

	// モックリザルト作成
	result := &ProcessResult{
		TotalArticles:    10,
		ProcessedCount:   8,
		ValidationErrors: 1,
		ProcessingErrors: 1,
		ErrorDetails:     []string{"Error 1", "Error 2"},
	}

	// CaptureOutputを使用
	output := CaptureOutput(func() {
		DisplayResults(result)
	})

	// 出力の検証
	assert.Contains(t, output, "変換結果:")
	assert.Contains(t, output, "8 記事が正常")
	// その他の検証...
}

// --- ReportMemoryUsage と ReportProcessResult のテスト ---

func TestReportMemoryUsage(t *testing.T) {
	dummy := &dummyReporter{}
	stage := "テスト段階"
	ReportMemoryUsage(dummy, stage)
	// 出力内容に stage とメモリ関連のキーワードが含まれていることを検証
	assert.Contains(t, dummy.Info, stage)
	assert.Contains(t, dummy.Info, "ヒープ=")
	assert.Contains(t, dummy.Info, "合計=")
	assert.Contains(t, dummy.Info, "システム=")
}

func TestReportProcessResult(t *testing.T) {
	dummy := &dummyReporter{}
	pr := NewProcessResult(5)
	ReportProcessResult(dummy, pr)
	assert.True(t, dummy.DisplayCalled, "DisplayResult が呼ばれていること")
	assert.Equal(t, pr, dummy.ReceivedResult, "正しい ProcessResult が渡されていること")
}
