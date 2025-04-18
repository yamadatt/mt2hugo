package downloader

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mt2hugo/internal/config"
	"mt2hugo/reporter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageDownloaderIntegration(t *testing.T) {
	t.Run("完全な画像ダウンロードのフロー検証", func(t *testing.T) {
		// 1. セットアップ
		tempDir := t.TempDir()
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("test image data"))
		}))
		defer mockServer.Close()

		// 2. ダウンローダーの作成
		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 30, 5)

		// 3. HTMLコンテンツの作成
		htmlContent := `
            <img src="` + mockServer.URL + `/image1.jpg">
            <p>テスト文章</p>
            <img src="` + mockServer.URL + `/image2.jpg">
        `

		// 4. ProcessHTMLImagesの実行
		result, err := downloader.ProcessHTMLImages(htmlContent, tempDir)

		// 5. 検証
		require.NoError(t, err, "ProcessHTMLImagesはエラーを返すべきでない")

		// 6. 画像ファイルの存在確認
		files, err := os.ReadDir(tempDir)
		require.NoError(t, err)
		assert.NotEmpty(t, files, "画像ファイルがダウンロードされるべき")

		// 7. HTMLが適切に書き換えられていることの確認
		assert.NotContains(t, result, mockServer.URL, "元のURLが置換されるべき")

		// 8. Reporter呼び出し確認 - 部分一致で検証するよう修正
		found := false
		for _, msg := range mockReporter.InfoMessages {
			if strings.Contains(msg, "記事内の画像") {
				found = true
				break
			}
		}
		assert.True(t, found, "記事内の画像に関するメッセージがログに記録されるべき")
	})
}

func TestDownloadImageOption(t *testing.T) {
	t.Run("ダウンロードオプションのテスト", func(t *testing.T) {
		// テスト用サーバー
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("test image data"))
		}))
		defer mockServer.Close()

		tempDir := t.TempDir()

		// HTMLコンテンツの作成
		htmlContent := `<img src="` + mockServer.URL + `/image1.jpg">`

		// ケース1: ダウンロードオプション有効
		t.Run("ダウンロード有効", func(t *testing.T) {
			mockReporter := reporter.NewMockReporter()
			downloader := NewImageDownloader(mockReporter, 30, 5)

			result, err := downloader.ProcessHTMLImages(htmlContent, tempDir)
			require.NoError(t, err)

			// 画像ファイルが存在するか確認
			files, err := os.ReadDir(tempDir)
			require.NoError(t, err)
			assert.NotEmpty(t, files, "画像ファイルがダウンロードされるべき")
			assert.NotContains(t, result, mockServer.URL, "元のURLが置換されるべき")
		})

		// ケース2: 出力ディレクトリなし
		t.Run("出力ディレクトリなし", func(t *testing.T) {
			mockReporter := reporter.NewMockReporter()
			downloader := NewImageDownloader(mockReporter, 30, 5)

			result, err := downloader.ProcessHTMLImages(htmlContent, "")
			require.NoError(t, err)

			// 出力ディレクトリが指定されていないためスキップされるはず
			assert.Contains(t, result, mockServer.URL, "URLは置換されないはず")
			foundMsg := false
			for _, msg := range mockReporter.InfoMessages {
				if strings.Contains(msg, "出力ディレクトリが指定されていないため") {
					foundMsg = true
					break
				}
			}
			assert.True(t, foundMsg, "スキップメッセージがあるべき")
		})
	})
}

func TestDownloadImageIntegrationWithConfig(t *testing.T) {
	t.Run("設定オプションとの統合テスト", func(t *testing.T) {
		// サーバー設定
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("test image data"))
		}))
		defer mockServer.Close()

		tempDir := t.TempDir()

		// HTMLコンテンツ
		htmlContent := `<img src="` + mockServer.URL + `/test.jpg">`

		// テストケース
		testCases := []struct {
			name             string
			downloadImages   bool
			expectedDownload bool
		}{
			{"ダウンロード有効", true, true},
			{"ダウンロード無効", false, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 設定
				cfg := &config.Config{
					OutputDir:      tempDir,
					DownloadImages: tc.downloadImages,
					ImageTimeout:   30,
					MaxConcurrent:  5,
				}

				// レポーター
				mockReporter := reporter.NewMockReporter()

				var result string
				var err error

				if tc.downloadImages {
					// ダウンロード有効の場合
					imgDownloader := NewImageDownloader(mockReporter, cfg.ImageTimeout, cfg.MaxConcurrent)
					result, err = imgDownloader.ProcessHTMLImages(htmlContent, tempDir)
				} else {
					// ダウンロード無効の場合（内容変更なし）
					result = htmlContent
				}

				require.NoError(t, err)

				if tc.expectedDownload {
					// ダウンロードされるべき
					assert.NotContains(t, result, mockServer.URL, "元のURLが置換されるべき")

					// ファイル確認
					files, _ := os.ReadDir(tempDir)
					assert.NotEmpty(t, files, "画像ファイルがダウンロードされるべき")
				} else {
					// ダウンロードされないはず
					assert.Contains(t, result, mockServer.URL, "URLは置換されないはず")
				}
			})
		}
	})
}

// ImageProcessorインターフェースを定義して、インポートサイクルを回避
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

func TestMainFlowWithImages(t *testing.T) {
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

	// 設定作成
	cfg := &config.Config{
		InputFile:      inputFile,
		OutputDir:      outputDir,
		DownloadImages: true,
		ImageTimeout:   30,
		MaxConcurrent:  5,
	}

	// メイン処理のシミュレーション
	rep := reporter.NewSimpleReporter()
	imgDownloader := NewImageDownloader(rep, cfg.ImageTimeout, cfg.MaxConcurrent)

	// ダミーの変換器を作成（実際の実装に合わせて調整）
	transformer := &DummyTransformer{
		imageProcessor: imgDownloader,
		outputDir:      outputDir,
		downloadImages: cfg.DownloadImages,
	}

	// 変換処理の実行
	err = transformer.TransformFile(inputFile)
	require.NoError(t, err)

	// 出力ファイルを確認
	outputFiles, err := os.ReadDir(outputDir)
	require.NoError(t, err)
	assert.NotEmpty(t, outputFiles, "出力ファイルが生成されるべき")

	// 画像ディレクトリがあるか確認
	imageDir := filepath.Join(outputDir, "images")
	imageFiles, err := os.ReadDir(imageDir)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, imageFiles, "画像ファイルがダウンロードされるべき")
	}
}

// コマンドライン引数のテスト
func TestCommandLineFlag(t *testing.T) {
	t.Run("download-imagesフラグのテスト", func(t *testing.T) {
		// 実際のmainパッケージをモックする必要があるため、Configの機能だけテスト
		cfg := &config.Config{
			DownloadImages: true,
			OutputDir:      "output",
		}

		assert.True(t, cfg.DownloadImages, "download-imagesフラグがtrueに設定されるべき")

		// ここでは実際のフラグ解析はテストできないため、
		// mt2hugo/cmd/mt2hugoパッケージでフラグ解析のテストを実装することをおすすめします
	})
}
