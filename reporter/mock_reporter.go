package reporter

import "fmt"

// MockReporter はテスト用のReporterインターフェース実装
type MockReporter struct {
	WarningMessages []string
	InfoMessages    []string
	Messages        []string
	ErrorMessages   []string
}

// NewMockReporter は新しいMockReporterを作成する
func NewMockReporter() *MockReporter {
	return &MockReporter{
		WarningMessages: []string{},
		InfoMessages:    []string{},
		Messages:        []string{},
		ErrorMessages:   []string{},
	}
}

// PrintWarning は警告メッセージを記録する
func (m *MockReporter) PrintWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.WarningMessages = append(m.WarningMessages, message)
}

// PrintInfo は情報メッセージを記録する
func (m *MockReporter) PrintInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.InfoMessages = append(m.InfoMessages, message)
}

// PrintMessage はメッセージを記録する
func (m *MockReporter) PrintMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.Messages = append(m.Messages, message)
}

// PrintError はエラーメッセージを記録する
func (m *MockReporter) PrintError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	m.ErrorMessages = append(m.ErrorMessages, message)
}

// Start は処理開始を記録する
func (m *MockReporter) Start(message string) {}

// UpdateProgress は進捗更新を記録する
func (m *MockReporter) UpdateProgress(current int, message string) {}

// Finish は処理終了を記録する
func (m *MockReporter) Finish(message string) {}

// DisplayResult は結果表示を記録する
func (m *MockReporter) DisplayResult(result *ProcessResult) {
	m.InfoMessages = append(m.InfoMessages, "処理結果を表示 (テスト用)")
}
