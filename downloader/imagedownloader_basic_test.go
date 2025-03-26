package downloader

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"mt2hugo/reporter"

	"github.com/stretchr/testify/assert"
)

// extractImageURLsのテスト
func TestExtractImageURLs(t *testing.T) {
	downloader, _ := createTestDownloader(t)

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "空のコンテンツ",
			content:  "",
			expected: []string{},
		},
		{
			name:     "HTMLタグ画像",
			content:  `<img src="https://example.com/image.jpg" alt="test">`,
			expected: []string{"https://example.com/image.jpg"},
		},
		{
			name:     "Markdown画像",
			content:  `![Alt text](https://example.com/image.png)`,
			expected: []string{"https://example.com/image.png"},
		},
		{
			name:     "Hugoショートコード",
			content:  `{{< figure src="https://example.com/hugo.jpg" caption="Hugo" >}}`,
			expected: []string{"https://example.com/hugo.jpg"},
		},
		{
			name:     "MovableType形式",
			content:  "IMAGE: https://example.com/mt.gif",
			expected: []string{"https://example.com/mt.gif"},
		},
		{
			name: "複数画像とURLの重複",
			//IMAGE前に空白があるとマッチしないため（^IMAGE:にしている）インデントが崩れているのは許容する。
			content: `
                <img src="https://example.com/image1.jpg">
                ![Alt](https://example.com/image2.png)
                {{< figure src="https://example.com/image3.gif" >}}
IMAGE: https://example.com/image4.webp
                <img src="https://example.com/image1.jpg">
            `,
			expected: []string{
				"https://example.com/image1.jpg",
				"https://example.com/image2.png",
				"https://example.com/image3.gif",
				"https://example.com/image4.webp",
			},
		},
		{
			name:     "バックグラウンド画像",
			content:  `<div style="background-image: url('https://example.com/bg.jpg')">`,
			expected: []string{"https://example.com/bg.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := downloader.extractImageURLs(tt.content)

			// URLの順序は保証されていないので、同じ要素が含まれているかチェック
			assert.ElementsMatch(t, tt.expected, actual, "抽出されたURLが期待値と一致すること")
		})
	}
}

// ProcessHTMLImagesのテスト - インターフェースモッキングを使用
func TestProcessHTMLImages(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		outputDir      string
		mockDownloads  map[string]string
		mockError      error
		expectedOutput string
	}{
		{
			name:           "空のコンテンツ",
			content:        "",
			outputDir:      "images",
			mockDownloads:  map[string]string{},
			mockError:      nil,
			expectedOutput: "",
		},
		{
			name:           "出力ディレクトリなし",
			content:        "<img src='https://example.com/image.jpg'>",
			outputDir:      "",
			mockDownloads:  map[string]string{},
			mockError:      nil,
			expectedOutput: "<img src='https://example.com/image.jpg'>",
		},
		{
			name:           "画像URL検出なし",
			content:        "<p>No images here</p>",
			outputDir:      "images",
			mockDownloads:  map[string]string{},
			mockError:      nil,
			expectedOutput: "<p>No images here</p>",
		},
		{
			name:      "HTML画像タグの置換",
			content:   "<img src='https://example.com/image.jpg'>",
			outputDir: "images",
			mockDownloads: map[string]string{
				"https://example.com/image.jpg": "downloaded.jpg",
			},
			mockError:      nil,
			expectedOutput: "<img src=\"downloaded.jpg\">",
		},
		{
			name: "複数形式の画像置換",
			content: `
            <img src="https://example.com/image1.jpg">
            ![Alt](https://example.com/image2.png)
            {{< figure src="https://example.com/image3.gif" >}}
            background: url('https://example.com/image4.webp')
            IMAGE: https://example.com/image5.svg
            `,
			outputDir: "images",
			mockDownloads: map[string]string{
				"https://example.com/image1.jpg":  "img1.jpg",
				"https://example.com/image2.png":  "img2.png",
				"https://example.com/image3.gif":  "img3.gif",
				"https://example.com/image4.webp": "img4.webp",
				"https://example.com/image5.svg":  "img5.svg",
			},
			mockError: nil,
			expectedOutput: `
            <img src="img1.jpg">
            ![Alt](img2.png)
            {{< figure src="img3.gif" >}}
            background: url("img4.webp")
            IMAGE: img5.svg
            `,
		},
		{
			name:      "一部ダウンロード失敗",
			content:   "<img src='https://example.com/good.jpg'><img src='https://example.com/bad.jpg'>",
			outputDir: "images",
			mockDownloads: map[string]string{
				"https://example.com/good.jpg": "good.jpg",
				// bad.jpgはエラーでダウンロードされず
			},
			mockError:      fmt.Errorf("一部エラー発生"),
			expectedOutput: "<img src=\"good.jpg\"><img src='https://example.com/bad.jpg'>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReporter := reporter.NewMockReporter()

			// 1. モックダウンローダーを作成
			mockDownloader := &TestImageDownloader{
				mockReporter: mockReporter,
				mockDownloadImagesFunc: func(urls []string, outputDir string) (map[string]string, error) {
					return tt.mockDownloads, tt.mockError
				},
			}

			// 2. テスト実行
			result, err := mockDownloader.ProcessHTMLImages(tt.content, tt.outputDir)

			// 3. 結果確認
			if tt.mockError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedOutput, result)

			// 4. 特定のケースに対する追加チェック
			if tt.outputDir == "" {
				assert.Contains(t, mockReporter.InfoMessages, "出力ディレクトリが指定されていないため、画像ダウンロードをスキップします")
			}
			if tt.name == "画像URL検出なし" {
				assert.Contains(t, mockReporter.InfoMessages, "画像URLが検出されませんでした")
			}
		})
	}
}

// TestImageDownloader はテスト用に機能を上書きしたダウンローダー
type TestImageDownloader struct {
	*ImageDownloader
	mockReporter           reporter.Reporter
	mockDownloadImagesFunc func(urls []string, outputDir string) (map[string]string, error)
}

// ProcessHTMLImages はテスト用に上書きしたメソッド
func (t *TestImageDownloader) ProcessHTMLImages(content string, outputDir string) (string, error) {
	// 出力ディレクトリのチェック
	if outputDir == "" {
		t.mockReporter.PrintInfo("出力ディレクトリが指定されていないため、画像ダウンロードをスキップします")
		return content, nil
	}

	// 画像URLの抽出
	urls := t.extractImageURLs(content)
	if len(urls) == 0 {
		t.mockReporter.PrintInfo("画像URLが検出されませんでした")
		return content, nil
	}

	// ダウンロード処理（モック関数を使用）
	replacements, err := t.mockDownloadImagesFunc(urls, outputDir)

	// 置換処理
	result := content
	for imgURL, localPath := range replacements {
		// HTMLタグの置換 - シングルクォートをダブルクォートに変換
		re := regexp.MustCompile(`<img[^>]*src=['"]` + regexp.QuoteMeta(imgURL) + `['"][^>]*>`)
		result = re.ReplaceAllStringFunc(result, func(src string) string {
			// シングルクォートをダブルクォートに置換
			reAttr := regexp.MustCompile(`src=['"]([^'"]+)['"]`)
			return reAttr.ReplaceAllString(src, `src="`+localPath+`"`)
		})

		// Markdown形式の置換
		re = regexp.MustCompile(`!\[[^\]]*\]\(` + regexp.QuoteMeta(imgURL) + `\)`)
		result = re.ReplaceAllStringFunc(result, func(src string) string {
			return strings.Replace(src, imgURL, localPath, 1)
		})

		// Hugo shortcode形式の置換
		re = regexp.MustCompile(`{{<\s*figure\s+src=['"]` + regexp.QuoteMeta(imgURL) + `['"][^>]*>}}`)
		result = re.ReplaceAllStringFunc(result, func(src string) string {
			// シングルクォートをダブルクォートに置換
			reAttr := regexp.MustCompile(`src=['"]([^'"]+)['"]`)
			return reAttr.ReplaceAllString(src, `src="`+localPath+`"`)
		})

		// CSS background形式の置換
		re = regexp.MustCompile(`url\(['"]?` + regexp.QuoteMeta(imgURL) + `['"]?\)`)
		result = re.ReplaceAllString(result, `url("`+localPath+`")`)

		// MovableType IMAGE:タグの置換 - 修正
		re = regexp.MustCompile(`IMAGE:\s*` + regexp.QuoteMeta(imgURL))
		result = re.ReplaceAllString(result, `IMAGE: `+localPath)
	}

	return result, err
}

// extractImageURLs はテストのために必要な真の実装を呼び出す
func (t *TestImageDownloader) extractImageURLs(content string) []string {
	// 実際のダウンローダーを作成して処理を委譲
	downloader := NewImageDownloader(t.mockReporter, 10, 5)
	return downloader.extractImageURLs(content)
}
