package downloader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mt2hugo/reporter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// downloadImageの実際のHTTP呼び出しを使用するテスト
func TestDownloadImage_RealHTTP(t *testing.T) {
	// テストサーバーのセットアップ
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("fake image data"))
		case "/notfound.jpg":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()

	// テスト用のレポーター
	mockReporter := reporter.NewMockReporter()

	// ダウンローダの作成
	downloader := NewImageDownloader(mockReporter, 10, 5)

	tests := []struct {
		name        string
		imgURL      string
		shouldError bool
		checkFile   bool
	}{
		{
			name:        "相対URL",
			imgURL:      "images/local.jpg",
			shouldError: false,
			checkFile:   false,
		},
		{
			name:        "不正なURL",
			imgURL:      "http://[::1]:invalid",
			shouldError: true,
			checkFile:   false,
		},
		{
			name:        "存在するリソース",
			imgURL:      server.URL + "/valid.jpg",
			shouldError: false,
			checkFile:   true,
		},
		{
			name:        "404エラー",
			imgURL:      server.URL + "/notfound.jpg",
			shouldError: true,
			checkFile:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用のダウンローダを使用
			result, err := downloader.downloadImage(tt.imgURL, tempDir)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tt.imgURL == "images/local.jpg" {
					// 相対URLの場合は元のURLが返される
					assert.Equal(t, tt.imgURL, result)
				}

				if tt.checkFile {
					// ファイルがダウンロードされたことを確認
					filePath := filepath.Join(tempDir, result)
					_, err := os.Stat(filePath)
					assert.NoError(t, err, "ファイルが存在すること")

					// ファイルの内容を確認
					data, err := ioutil.ReadFile(filePath)
					assert.NoError(t, err)
					assert.Equal(t, "fake image data", string(data))
				}
			}
		})
	}
}

// モックを使用するテスト - インターフェースを活用
func TestDownloadImage_WithMock(t *testing.T) {

	tests := []struct {
		name        string
		imgURL      string
		outputDir   string
		mockPath    string
		mockError   error
		expectError bool
	}{
		{
			name:        "成功するケース",
			imgURL:      "https://example.com/image.jpg",
			outputDir:   "images",
			mockPath:    "downloaded.jpg",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "エラーが発生するケース",
			imgURL:      "https://error.example.com/image.jpg",
			outputDir:   "images",
			mockPath:    "",
			mockError:   fmt.Errorf("ダウンロードエラー"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. モックレポーターとダウンローダーを作成
			mockReporter := reporter.NewMockReporter()
			downloader := NewImageDownloader(mockReporter, 10, 5)

			// 2. モックダウンローダーを作成 (正しく実装)
			mockDownloader := &MockImageDownloader{
				mockDownloadImageFunc: func(imgURL string, outputDir string) (string, error) {
					// テストケースの値を返す
					// 実際のHTTP呼び出しは行わない
					return tt.mockPath, tt.mockError
				},
			}

			// 3. モックをセット - デバッグのためにログを追加
			downloader.SetDownloader(mockDownloader)

			// 4. モック関数が確実に使われるよう、ファイルを作成
			if tt.mockPath != "" && tt.mockError == nil {
				fullPath := filepath.Join(tt.outputDir, tt.mockPath)
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				require.NoError(t, err, "テスト用ディレクトリの作成に失敗")

				// 空ファイルを作成
				f, err := os.Create(fullPath)
				require.NoError(t, err, "テスト用ファイルの作成に失敗")
				f.Close()
			}

			// 5. テスト対象メソッドを呼び出し
			result, err := downloader.DownloadImage(tt.imgURL, tt.outputDir)

			// 6. 結果を検証
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockPath, result)
			}
		})
	}
}

// downloadImagesのテスト - 並行ダウンロード処理のモック
func TestDownloadImages_WithMock(t *testing.T) {
	tempDir := t.TempDir()

	// テスト用のURLリスト
	urls := []string{
		"https://example.com/image1.jpg",
		"https://example.com/image2.png",
		"https://example.com/image3.gif",
	}

	// ダウンロード成功のケース
	t.Run("全ダウンロード成功", func(t *testing.T) {
		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 10, 5)

		// 呼び出し回数をカウント
		downloadCount := 0

		// モックダウンローダを作成
		mockDownloader := &MockImageDownloader{
			mockDownloadImageFunc: func(imgURL, outputDir string) (string, error) {
				// 呼び出し回数をインクリメント
				downloadCount++

				// URLごとに異なる結果を返す
				switch imgURL {
				case "https://example.com/image1.jpg":
					return "downloaded1.jpg", nil
				case "https://example.com/image2.png":
					return "downloaded2.png", nil
				case "https://example.com/image3.gif":
					return "downloaded3.gif", nil
				default:
					return "", fmt.Errorf("未知のURL: %s", imgURL)
				}
			},
		}

		// モックをセット
		downloader.SetDownloader(mockDownloader)

		// ダウンロードを実行
		result, err := downloader.downloadImages(urls, tempDir)

		// 結果を検証
		assert.NoError(t, err)
		assert.Len(t, result, 3, "3つのファイルがダウンロードされるべき")
		assert.Equal(t, 3, downloadCount, "downloadImageが3回呼ばれるべき")

		// 期待される結果
		assert.Equal(t, "downloaded1.jpg", result["https://example.com/image1.jpg"])
		assert.Equal(t, "downloaded2.png", result["https://example.com/image2.png"])
		assert.Equal(t, "downloaded3.gif", result["https://example.com/image3.gif"])
	})

	// 一部エラーのケース
	t.Run("一部ダウンロード失敗", func(t *testing.T) {
		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 10, 5)

		// モックダウンローダを作成
		mockDownloader := &MockImageDownloader{
			mockDownloadImageFunc: func(imgURL, outputDir string) (string, error) {
				// 特定のURLでエラーを返す
				if imgURL == "https://example.com/image2.png" {
					return "", fmt.Errorf("ダウンロードエラー")
				}

				// それ以外は成功
				switch imgURL {
				case "https://example.com/image1.jpg":
					return "downloaded1.jpg", nil
				case "https://example.com/image3.gif":
					return "downloaded3.gif", nil
				default:
					return "", fmt.Errorf("未知のURL: %s", imgURL)
				}
			},
		}

		// モックをセット
		downloader.SetDownloader(mockDownloader)

		// ダウンロードを実行
		result, err := downloader.downloadImages(urls, tempDir)

		// 結果を検証
		assert.Error(t, err, "エラーが返されるべき")
		assert.Contains(t, err.Error(), "ダウンロードエラー")

		// 失敗しても成功した結果はマップに含まれる
		assert.Len(t, result, 2, "2つのダウンロードは成功する")
		assert.Equal(t, "downloaded1.jpg", result["https://example.com/image1.jpg"])
		assert.Equal(t, "downloaded3.gif", result["https://example.com/image3.gif"])
	})

	// キャッシュ使用のケース
	t.Run("キャッシュ利用", func(t *testing.T) {
		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 10, 5)

		// キャッシュをセットアップ
		downloader.mutex.Lock()
		downloader.downloadCache["https://example.com/image1.jpg"] = "cached1.jpg"
		downloader.downloadCache["https://example.com/image2.png"] = "cached2.png"
		downloader.mutex.Unlock()

		// キャッシュに対応するファイルを作成
		for _, path := range []string{"cached1.jpg", "cached2.png"} {
			fullPath := filepath.Join(tempDir, path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			require.NoError(t, err)
			err = ioutil.WriteFile(fullPath, []byte("cached content"), 0644)
			require.NoError(t, err)
		}

		// モックダウンローダを作成 - キャッシュミスの場合のみ呼ばれる
		mockDownloader := &MockImageDownloader{
			mockDownloadImageFunc: func(imgURL, outputDir string) (string, error) {
				// image3のみダウンロードが必要
				if imgURL == "https://example.com/image3.gif" {
					return "downloaded3.gif", nil
				}

				// それ以外（キャッシュヒット）は呼ばれないはず
				t.Errorf("キャッシュがあるためダウンロードは不要: %s", imgURL)
				return "", fmt.Errorf("should not be called")
			},
		}

		// モックをセット
		downloader.SetDownloader(mockDownloader)

		// ダウンロードを実行
		result, err := downloader.downloadImages(urls, tempDir)

		// 結果を検証
		assert.NoError(t, err)
		assert.Len(t, result, 3, "3つのファイルが取得されるべき")

		// 期待される結果
		assert.Equal(t, "cached1.jpg", result["https://example.com/image1.jpg"])
		assert.Equal(t, "cached2.png", result["https://example.com/image2.png"])
		assert.Equal(t, "downloaded3.gif", result["https://example.com/image3.gif"])
	})
}
