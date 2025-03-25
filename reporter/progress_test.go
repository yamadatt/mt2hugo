package reporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressReporter_Start(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		total    int
		expected string
	}{
		{
			name:     "基本の開始メッセージ",
			message:  "開始",
			total:    10,
			expected: "開始...\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := NewProgressReporterWithWriter(tc.total, &buf)
			pr.Start(tc.message)
			assert.Equal(t, tc.expected, buf.String(), "Start の出力が期待通りであること")
		})
	}
}

func TestProgressReporter_UpdateProgress(t *testing.T) {
	tests := []struct {
		name               string
		current            int
		total              int
		lastUpdateDelta    time.Duration // 現在時刻からの相対値（負なら過去に設定）
		message            string
		expectOutput       bool
		expectedSubstrings []string
	}{
		{
			name:               "十分な間隔がある場合は出力される",
			current:            5,
			total:              10,
			lastUpdateDelta:    -time.Second,
			message:            "進捗",
			expectOutput:       true,
			expectedSubstrings: []string{"進捗", "5/10", "50.0%"},
		},
		{
			name:            "更新間隔が短い場合は出力されない",
			current:         6,
			total:           10,
			lastUpdateDelta: 0, // 現在時刻をセット
			message:         "進捗",
			expectOutput:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := NewProgressReporterWithWriter(tc.total, &buf)
			// lastUpdate をテスト用に設定
			pr.lastUpdate = time.Now().Add(tc.lastUpdateDelta)
			pr.UpdateProgress(tc.current, tc.message)
			output := buf.String()

			if tc.expectOutput {
				for _, substr := range tc.expectedSubstrings {
					assert.Contains(t, output, substr, "出力に %q が含まれること", substr)
				}
			} else {
				assert.Empty(t, output, "更新間隔が短い場合は何も出力されない")
			}
		})
	}
}

func TestProgressReporter_Finish(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		total    int
		expected string
	}{
		{
			name:     "基本の終了メッセージ",
			message:  "終了",
			total:    10,
			expected: "\n終了\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			pr := NewProgressReporterWithWriter(tc.total, &buf)
			pr.Finish(tc.message)
			assert.Equal(t, tc.expected, buf.String(), "Finish の出力が期待通りであること")
		})
	}
}

func TestProgressReporter_PrintMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   string // "message", "warning", "info"
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "PrintMessage の出力",
			method:   "message",
			format:   "Hello %s",
			args:     []interface{}{"World"},
			expected: "Hello World\n",
		},
		{
			name:     "PrintWarning の出力",
			method:   "warning",
			format:   "Warning: %s",
			args:     []interface{}{"Test"},
			expected: "警告: Warning: Test\n",
		},
		{
			name:     "PrintInfo の出力",
			method:   "info",
			format:   "Info: %s",
			args:     []interface{}{"Test"},
			expected: "情報: Info: Test\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr := &ProgressReporter{}
			output := CaptureOutput(func() { // 修正: captureOutput -> CaptureOutput
				switch tc.method {
				case "message":
					pr.PrintMessage(tc.format, tc.args...)
				case "warning":
					pr.PrintWarning(tc.format, tc.args...)
				case "info":
					pr.PrintInfo(tc.format, tc.args...)
				}
			})
			assert.Equal(t, tc.expected, output, "%s の出力が期待通りであること", tc.method)
		})
	}
}

func TestConsoleReporter_PrintMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   string // "message", "info"
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "ConsoleReporter PrintMessage の出力",
			method:   "message",
			format:   "Console %s",
			args:     []interface{}{"Message"},
			expected: "Console Message\n",
		},
		{
			name:     "ConsoleReporter PrintInfo の出力",
			method:   "info",
			format:   "Console %s",
			args:     []interface{}{"Info"},
			expected: "情報: Console Info\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cr := &ConsoleReporter{}
			output := CaptureOutput(func() { // 修正: captureOutput -> CaptureOutput
				switch tc.method {
				case "message":
					cr.PrintMessage(tc.format, tc.args...)
				case "info":
					cr.PrintInfo(tc.format, tc.args...)
				}
			})
			assert.Equal(t, tc.expected, output, "ConsoleReporter の %s 出力が期待通りであること", tc.method)
		})
	}
}
