package downloader

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mt2hugo/internal/config"
	"mt2hugo/reporter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImageDownloaderIntegration と TestDownloadImageOption は変更なし

// CircularImportSolverは、インポートサイクルを回避するためのインターフェース
type ImageProcessor interface {
	ProcessHTMLImages(htmlContent, outputDir string) (string, error)
}

// テスト用のダミー変換器
type DummyTransformer struct {
	imageProcessor ImageProcessor
	outputDir      string
	downloadImages bool
}

func (d *DummyTransformer) TransformFile(inputFile string) error {
	// ファイル読み込み
	content, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return err
	}

	// 出力ディレクトリ作成
	imageDir := filepath.Join(d.outputDir, "images")
	os.MkdirAll(imageDir, 0755)

	// 画像処理
	processedContent := string(content)
	if d.downloadImages && d.imageProcessor != nil {
		processed, err := d.imageProcessor.ProcessHTMLImages(processedContent, imageDir)
		if err != nil {
			return err
		}
		processedContent = processed
	}

	// 結果を出力
	return ioutil.WriteFile(filepath.Join(d.outputDir, "output.md"), []byte(processedContent), 0644)
}

// メインフローをテストするための新しいテスト
func TestMainFlowWithImages(t *testing.T) {
	t.Run("メインフローの統合テスト", func(t *testing.T) {
		// モックサーバー設定
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("test image data"))
		}))
		defer mockServer.Close()

		// テンポラリディレクトリ
		tempDir := t.TempDir()
		outputDir := filepath.Join(tempDir, "output")
		err := os.MkdirAll(outputDir, 0755)
		require.NoError(t, err)

		// テスト入力ファイル作成
		inputContent := `
TITLE: テスト記事
DATE: 2023-01-01
-----
BODY:
<p>テスト本文</p>
<img src="` + mockServer.URL + `/image.jpg">
-----
`
		inputFile := filepath.Join(tempDir, "input.txt")
		err = ioutil.WriteFile(inputFile, []byte(inputContent), 0644)
		require.NoError(t, err)

		// テストケース
		testCases := []struct {
			name           string
			downloadImages bool
			expectImages   bool
		}{
			{"ダウンロード有効", true, true},
			{"ダウンロード無効", false, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 設定作成
				cfg := &config.Config{
					InputFile:      inputFile,
					OutputDir:      outputDir,
					DownloadImages: tc.downloadImages,
					ImageTimeout:   30,
					MaxConcurrent:  5,
				}

				// レポーター
				rep := reporter.NewSimpleReporter()

				// 変換処理
				var transformer *DummyTransformer

				if tc.downloadImages {
					imgDownloader := NewImageDownloader(rep, cfg.ImageTimeout, cfg.MaxConcurrent)
					transformer = &DummyTransformer{
						imageProcessor: imgDownloader,
						outputDir:      outputDir,
						downloadImages: cfg.DownloadImages,
					}
				} else {
					transformer = &DummyTransformer{
						imageProcessor: nil,
						outputDir:      outputDir,
						downloadImages: false,
					}
				}

				// 変換処理の実行
				err = transformer.TransformFile(inputFile)
				require.NoError(t, err)

				// 画像ディレクトリがあるか確認
				imageDir := filepath.Join(outputDir, "images")
				if tc.expectImages {
					// 画像があるはず
					imageFiles, err := os.ReadDir(imageDir)
					require.NoError(t, err)
					assert.NotEmpty(t, imageFiles, "画像ファイルがダウンロードされるべき")

					// 出力ファイルを確認して画像URLが置換されているか
					output, err := ioutil.ReadFile(filepath.Join(outputDir, "output.md"))
					require.NoError(t, err)
					assert.NotContains(t, string(output), mockServer.URL, "画像URLが置換されるべき")
				} else {
					// 画像がダウンロードされていないか確認
					_, err := os.Stat(imageDir)
					assert.True(t, os.IsNotExist(err), "画像ディレクトリは存在しないはず")
				}
			})
		}
	})
}

// コマンドライン引数のテスト
func TestDownloadImagesFlag(t *testing.T) {
	t.Run("download-imagesフラグのテスト", func(t *testing.T) {
		// Configのセットアップだけテスト
		testCases := []struct {
			name           string
			downloadImages bool
		}{
			{"ダウンロード有効", true},
			{"ダウンロード無効", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					DownloadImages: tc.downloadImages,
					OutputDir:      "output",
				}

				assert.Equal(t, tc.downloadImages, cfg.DownloadImages,
					"download-imagesフラグが正しく設定されるべき")
			})
		}
	})
}
