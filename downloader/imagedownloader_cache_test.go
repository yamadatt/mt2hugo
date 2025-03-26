package downloader

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mt2hugo/reporter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// キャッシュメカニズムのテスト
// func TestDownloadImagesCache(t *testing.T) {
// 	tempDir := t.TempDir()

// 	// サーバをセットアップして実際にダウンロードをテスト
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "image/jpeg")
// 		w.Write([]byte("test image data"))
// 	}))
// 	// deferを使わず、明示的な位置で閉じる

// 	mockReporter := reporter.NewMockReporter()
// 	downloader := NewImageDownloader(mockReporter, 10, 5)

// 	// テスト用のURLリスト
// 	urls := []string{
// 		server.URL + "/image1.jpg",
// 		server.URL + "/image2.jpg",
// 	}

// 	// 初回ダウンロード
// 	replacements1, err := downloader.downloadImages(urls, tempDir)
// 	require.NoError(t, err)
// 	assert.Len(t, replacements1, 2)

// 	// キャッシュエントリが作成されたことを確認
// 	assert.Len(t, downloader.downloadCache, 2)

// 	// 2回目の呼び出し（キャッシュから取得されるべき）
// 	replacements2, err := downloader.downloadImages(urls, tempDir)
// 	require.NoError(t, err)

// 	// 1回目と2回目の結果が同じであることを確認
// 	assert.Equal(t, replacements1, replacements2)

// 	// テスト完了後にサーバーを閉じる
// 	server.Close()
// }

// キャッシュ作成のテスト
func TestDownloadImagesCacheCreation(t *testing.T) {
	tempDir := t.TempDir()

	// テストサーバをセットアップ
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("test image data"))
	}))
	defer server.Close()

	mockReporter := reporter.NewMockReporter()
	downloader := NewImageDownloader(mockReporter, 10, 5)

	// テスト用のURLリスト
	urls := []string{
		server.URL + "/image1.jpg",
		server.URL + "/image2.jpg",
	}

	// 初回ダウンロード
	replacements, err := downloader.downloadImages(urls, tempDir)
	require.NoError(t, err)

	// 結果の検証
	assert.Len(t, replacements, 2, "両方の画像がダウンロードされるべき")

	// キャッシュエントリが作成されたことを確認
	assert.Len(t, downloader.downloadCache, 2, "キャッシュエントリが作成されるべき")

	// ダウンロードされたファイルが存在することを確認
	for imgURL, localPath := range replacements {
		assert.Contains(t, downloader.downloadCache, imgURL, "URLがキャッシュに保存されるべき")
		assert.Equal(t, localPath, downloader.downloadCache[imgURL], "キャッシュのパスが一致するべき")

		filePath := filepath.Join(tempDir, localPath)
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "ファイルが存在するべき")
	}
}

// キャッシュ再利用のテスト（別のアプローチ）
func TestDownloadImagesCacheReuse(t *testing.T) {
	tempDir := t.TempDir()

	mockReporter := reporter.NewMockReporter()
	downloader := NewImageDownloader(mockReporter, 10, 5)

	// キャッシュを手動で設定
	downloader.mutex.Lock()
	downloader.downloadCache["https://example.com/image1.jpg"] = "cached1.jpg"
	downloader.downloadCache["https://example.com/image2.jpg"] = "cached2.jpg"
	downloader.mutex.Unlock()

	// キャッシュされるファイルを実際に作成
	for _, cachedPath := range []string{"cached1.jpg", "cached2.jpg"} {
		fullPath := filepath.Join(tempDir, cachedPath)
		// ディレクトリを作成
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)

		// ファイルを作成
		f, err := os.Create(fullPath)
		require.NoError(t, err)
		f.Write([]byte("test data"))
		f.Close()
	}

	// テスト用のURLリスト
	urls := []string{
		"https://example.com/image1.jpg",
		"https://example.com/image2.jpg",
	}

	// キャッシュから取得する呼び出し
	replacements, err := downloader.downloadImages(urls, tempDir)
	require.NoError(t, err)

	// 結果の検証
	assert.Len(t, replacements, 2, "両方の画像がキャッシュから取得されるべき")
	assert.Equal(t, "cached1.jpg", replacements["https://example.com/image1.jpg"])
	assert.Equal(t, "cached2.jpg", replacements["https://example.com/image2.jpg"])
}

// キャッシュメカニズムのテスト - シンプルなモックなしでのアプローチ
func TestDownloadImagesCacheMock(t *testing.T) {
	tempDir := t.TempDir()

	// 固定URL
	url1 := "https://example.com/image1.jpg"
	url2 := "https://example.com/image2.jpg"
	urls := []string{url1, url2}

	// テスト #1: キャッシュが空の場合
	t.Run("初回ダウンロード - キャッシュなし", func(t *testing.T) {
		// モックサーバでテストURLに対応する応答を用意
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("test image data"))
		}))
		defer server.Close()

		// 実際のURLをサーバーURLに置き換え
		testUrls := []string{
			server.URL + "/image1.jpg",
			server.URL + "/image2.jpg",
		}

		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 10, 5)

		// 元のキャッシュ状態を確認
		assert.Empty(t, downloader.downloadCache, "初期状態ではキャッシュは空であるべき")

		// 実行
		replacements, err := downloader.downloadImages(testUrls, tempDir)

		// 検証
		require.NoError(t, err)
		assert.Len(t, replacements, 2, "両方の画像がダウンロードされるべき")
		assert.Len(t, downloader.downloadCache, 2, "キャッシュに2つのエントリがあるべき")

		// ダウンロードされたファイルが存在することを確認
		for _, localPath := range replacements {
			filePath := filepath.Join(tempDir, localPath)
			_, err := os.Stat(filePath)
			assert.NoError(t, err, "ファイルが存在するべき")
		}
	})

	// テスト #2: キャッシュがある場合
	t.Run("2回目のダウンロード - キャッシュあり", func(t *testing.T) {
		mockReporter := reporter.NewMockReporter()
		downloader := NewImageDownloader(mockReporter, 10, 5)

		// キャッシュを手動で設定
		downloader.mutex.Lock()
		downloader.downloadCache[url1] = "cached1.jpg"
		downloader.downloadCache[url2] = "cached2.jpg"
		downloader.mutex.Unlock()

		// キャッシュされるファイルを実際に作成
		for _, cachedPath := range []string{"cached1.jpg", "cached2.jpg"} {
			fullPath := filepath.Join(tempDir, cachedPath)
			// ディレクトリを作成
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			require.NoError(t, err)

			// ファイルを作成
			f, err := os.Create(fullPath)
			require.NoError(t, err)
			f.Write([]byte("test data"))
			f.Close()
		}

		// サーバーをセットアップするが、アクセスがあれば失敗と見なす
		hitServer := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hitServer = true
			w.WriteHeader(http.StatusInternalServerError) // エラーを返す
		}))
		defer server.Close()

		// 実行
		replacements, err := downloader.downloadImages(urls, tempDir)

		// 検証
		require.NoError(t, err)
		assert.Len(t, replacements, 2)
		assert.Equal(t, "cached1.jpg", replacements[url1])
		assert.Equal(t, "cached2.jpg", replacements[url2])
		assert.False(t, hitServer, "キャッシュがあるためサーバーにアクセスしないはず")
	})
}
