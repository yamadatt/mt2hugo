package downloader

import (
	"fmt"
	"testing"

	"mt2hugo/reporter"
)

// テスト用のDownloaderを作成
func createTestDownloader(t *testing.T) (*ImageDownloader, *reporter.MockReporter) {
	mockReporter := reporter.NewMockReporter()
	return NewImageDownloader(mockReporter, 10, 5), mockReporter
}

// MockImageDownloader はテスト用のモックダウンローダー
type MockImageDownloader struct {
	mockDownloadImageFunc func(imgURL string, outputDir string) (string, error)
}

// DownloadImage インターフェースの実装（大文字に注意）
func (m *MockImageDownloader) DownloadImage(imgURL string, outputDir string) (string, error) {
	if m.mockDownloadImageFunc != nil {
		return m.mockDownloadImageFunc(imgURL, outputDir)
	}
	return "", fmt.Errorf("モック関数が設定されていません")
}

// 元のdownloadImages関数を保持するための変数
var downloadImages = func(d *ImageDownloader, urls []string, outputDir string) (map[string]string, error) {
	return d.downloadImages(urls, outputDir)
}
