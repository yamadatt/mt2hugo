package downloader

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"mt2hugo/reporter"

	"github.com/imroc/req/v3"
)

// ImageDownloaderInterface はダウンロード処理のインターフェース
type ImageDownloaderInterface interface {
	DownloadImage(imgURL string, outputDir string) (string, error)
}

// ImageDownloader は画像をダウンロードするための構造体
type ImageDownloader struct {
	reporter      reporter.Reporter
	client        *req.Client
	timeout       time.Duration
	maxConcurrent int
	downloadCache map[string]string // URL -> ローカルパス
	mutex         sync.Mutex
	downloader    ImageDownloaderInterface // 追加：実際のダウンロード処理を行うインターフェース
}

// NewImageDownloader は新しいImageDownloaderを作成する
func NewImageDownloader(reporter reporter.Reporter, timeoutSec int, maxConcurrent int) *ImageDownloader {
	client := req.C().
		SetTimeout(time.Duration(timeoutSec) * time.Second).
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	downloader := &ImageDownloader{
		reporter:      reporter,
		client:        client,
		timeout:       time.Duration(timeoutSec) * time.Second,
		maxConcurrent: maxConcurrent,
		downloadCache: make(map[string]string),
	}

	// 自分自身をダウンローダーとして設定
	downloader.downloader = downloader

	return downloader
}

// ProcessHTMLImages はHTML内の画像を検出してダウンロードし、パスを書き換えたHTMLを返す
func (d *ImageDownloader) ProcessHTMLImages(content string, outputDir string) (string, error) {
	if content == "" {
		return content, nil
	}

	// 出力ディレクトリが空の場合はダウンロードをスキップ
	if outputDir == "" {
		d.reporter.PrintInfo("出力ディレクトリが指定されていないため、画像ダウンロードをスキップします")
		return content, nil
	}

	// 画像URLを検出
	imgURLs := d.extractImageURLs(content)
	if len(imgURLs) == 0 {
		d.reporter.PrintInfo("画像URLが検出されませんでした")
		return content, nil
	}

	d.reporter.PrintInfo("記事内の画像 %d 件をダウンロードします", len(imgURLs))

	// 画像を並行ダウンロード
	replacements, err := d.downloadImages(imgURLs, outputDir)
	if err != nil {
		// エラーがあっても処理は続行し、成功した画像のみ置換する
		d.reporter.PrintWarning("一部の画像ダウンロードに失敗しました: %v", err)
	}

	// HTMLとMarkdownのリンクを置き換え
	processedContent := content
	for oldURL, newPath := range replacements {
		escapedURL := regexp.QuoteMeta(oldURL)

		// 1. HTML img タグの src 属性
		srcRe := regexp.MustCompile(`src\s*=\s*["']` + escapedURL + `["']`)
		processedContent = srcRe.ReplaceAllString(processedContent, `src="`+newPath+`"`)

		// 2. Markdown形式の画像リンク
		mdRe := regexp.MustCompile(`!\[([^\]]*)\]\(` + escapedURL + `\)`)
		processedContent = mdRe.ReplaceAllString(processedContent, `![$1](`+newPath+`)`)

		// 3. Hugoショートコード形式の画像URL（修正版）
		hugoRe := regexp.MustCompile(`({{<\s*figure\s+src=)["']` + escapedURL + `["']`)
		processedContent = hugoRe.ReplaceAllString(processedContent, `$1"`+newPath+`"`)

		// 4. バックグラウンド画像（CSSスタイル内）
		bgRe := regexp.MustCompile(`background(-image)?\s*:\s*url\(['"]*` + escapedURL + `['"]*\)`)
		processedContent = bgRe.ReplaceAllString(processedContent, `background$1: url("`+newPath+`")`)

		// 5. MovableType形式の画像フィールド
		mtRe := regexp.MustCompile(`(?m)^IMAGE:\s*` + escapedURL + `\s*$`)
		processedContent = mtRe.ReplaceAllString(processedContent, `IMAGE: `+newPath)
	}

	return processedContent, nil
}

// extractImageURLs はHTML内の画像URLを抽出する
func (d *ImageDownloader) extractImageURLs(content string) []string {
	urlMap := make(map[string]bool) // 重複を防止

	// 1. HTML形式の画像タグを検出
	imgRe := regexp.MustCompile(`<img[^>]*?src\s*=\s*["']?([^"'>\s]+)["']?[^>]*?>`)
	imgMatches := imgRe.FindAllStringSubmatch(content, -1)

	for _, match := range imgMatches {
		if len(match) >= 2 && match[1] != "" {
			cleanURL := strings.TrimSpace(match[1])
			urlMap[cleanURL] = true

			if d.reporter != nil {
				d.reporter.PrintInfo("HTML画像URLを検出: %s", cleanURL)
			}
		}
	}

	// 2. Markdown形式の画像リンクを検出
	mdImgRe := regexp.MustCompile(`!\[(?:[^\]]*)\]\(([^)]+)\)`)
	mdMatches := mdImgRe.FindAllStringSubmatch(content, -1)

	for _, match := range mdMatches {
		if len(match) >= 2 && match[1] != "" {
			cleanURL := strings.TrimSpace(match[1])
			urlMap[cleanURL] = true

			if d.reporter != nil {
				d.reporter.PrintInfo("Markdown画像URLを検出: %s", cleanURL)
			}
		}
	}

	// 3. Hugoショートコード形式の画像URLを検出
	hugoRe := regexp.MustCompile(`{{<\s*figure\s+src=["']?([^"'\s>]+)["']?[^>]*?>}}`)
	hugoMatches := hugoRe.FindAllStringSubmatch(content, -1)

	for _, match := range hugoMatches {
		if len(match) >= 2 && match[1] != "" {
			cleanURL := strings.TrimSpace(match[1])
			urlMap[cleanURL] = true

			if d.reporter != nil {
				d.reporter.PrintInfo("Hugoショートコード画像URLを検出: %s", cleanURL)
			}
		}
	}

	// 4. MovableType形式の画像フィールドを検出（既存のコード）
	mtImgRe := regexp.MustCompile(`(?m)^IMAGE:\s*(.+)$`)
	mtMatches := mtImgRe.FindAllStringSubmatch(content, -1)

	for _, match := range mtMatches {
		if len(match) >= 2 && match[1] != "" {
			cleanURL := strings.TrimSpace(match[1])
			if cleanURL != "" {
				urlMap[cleanURL] = true

				if d.reporter != nil {
					d.reporter.PrintInfo("MovableType画像URLを検出: %s", cleanURL)
				}
			}
		}
	}

	// バックグラウンド画像URLを検出
	bgRe := regexp.MustCompile(`background(-image)?\s*:\s*url\s*\(['"]?([^'")]+)['"]?\)`)
	bgMatches := bgRe.FindAllStringSubmatch(content, -1)

	for _, match := range bgMatches {
		if len(match) >= 3 && match[2] != "" {
			cleanURL := strings.TrimSpace(match[2])
			urlMap[cleanURL] = true

			if d.reporter != nil {
				d.reporter.PrintInfo("バックグラウンド画像URLを検出: %s", cleanURL)
			}
		}
	}

	// マップからスライスに変換
	var urls []string
	for url := range urlMap {
		urls = append(urls, url)
	}

	return urls
}

// downloadImages は画像を並行ダウンロードする
func (d *ImageDownloader) downloadImages(urls []string, outputDir string) (map[string]string, error) {
	replacements := make(map[string]string)
	var wg sync.WaitGroup
	sem := make(chan struct{}, d.maxConcurrent) // 同時実行数を制限
	var errorsMu sync.Mutex
	var downloadErrors []error

	// ディレクトリ作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("ディレクトリ作成に失敗しました: %s - %v", outputDir, err)
	}

	// 並行ダウンロード
	for _, imgURL := range urls {
		// キャッシュチェック
		d.mutex.Lock()
		if localPath, ok := d.downloadCache[imgURL]; ok {
			replacements[imgURL] = localPath
			d.mutex.Unlock()
			continue
		}
		d.mutex.Unlock()

		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			sem <- struct{}{}        // セマフォ取得
			defer func() { <-sem }() // セマフォ解放

			localPath, err := d.downloader.DownloadImage(url, outputDir)
			if err != nil {
				errorsMu.Lock()
				downloadErrors = append(downloadErrors, fmt.Errorf("URL: %s - %v", url, err))
				errorsMu.Unlock()
				return
			}

			d.mutex.Lock()
			replacements[url] = localPath
			d.downloadCache[url] = localPath
			d.mutex.Unlock()
		}(imgURL)
	}

	wg.Wait()

	// エラー処理
	if len(downloadErrors) > 0 {
		// 最初のエラーを報告するが、置換マップも返す
		return replacements, downloadErrors[0]
	}

	return replacements, nil
}

// DownloadImage は1つの画像をダウンロードする（公開インターフェースメソッド）
func (d *ImageDownloader) DownloadImage(imgURL string, outputDir string) (string, error) {
	// 元のprivateメソッドと同じ実装
	return d.downloadImage(imgURL, outputDir)
}

// SetDownloader はダウンロード処理のインターフェースを設定する
func (d *ImageDownloader) SetDownloader(downloader ImageDownloaderInterface) {
	d.downloader = downloader
}

// downloadImage は1つの画像をダウンロードする（内部実装）
func (d *ImageDownloader) downloadImage(imgURL string, outputDir string) (string, error) {
	// URLパース
	parsedURL, err := url.Parse(imgURL)
	if err != nil {
		d.reporter.PrintWarning("URLパースエラー: %s - %v", imgURL, err)
		return "", fmt.Errorf("不正なURL形式: %v", err)
	}

	// 相対URLの場合はスキップ
	if !parsedURL.IsAbs() {
		d.reporter.PrintInfo("相対URL - スキップ: %s", imgURL)
		return imgURL, nil // 相対パスはそのまま返す
	}

	// URL修正の試み - ?から始まるクエリ部分を削除
	cleanURL := imgURL
	if idx := strings.Index(cleanURL, "?"); idx > 0 {
		cleanURL = cleanURL[:idx]
		d.reporter.PrintInfo("URLクリーニング: %s -> %s", imgURL, cleanURL)
	}

	// ファイル名の取得
	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		// URLからファイル名を取得できない場合はタイムスタンプベースで生成
		fileName = fmt.Sprintf("image_%d%s", time.Now().UnixNano(), guessImageExt(imgURL))
		d.reporter.PrintInfo("ファイル名生成: %s", fileName)
	}

	// 保存先パス
	destPath := filepath.Join(outputDir, fileName)

	// ファイルが既に存在するか確認
	if fileInfo, err := os.Stat(destPath); err == nil && fileInfo.Size() > 0 {
		d.reporter.PrintInfo("ファイルが既に存在するためスキップ: %s", destPath)
		return fileName, nil // ダウンロードせずに既存のファイル名を返す
	}

	d.reporter.PrintInfo("ダウンロード開始: %s -> %s", imgURL, destPath)

	// imroc/reqを使用してダウンロード - エラーハンドリング強化
	client := d.client.R().
		SetOutputFile(destPath).
		EnableTrace().
		SetRetryCount(3) // リトライを追加

	// ユーザーエージェントとリファラーを追加
	client = client.SetHeader("Referer", fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host))

	resp, err := client.Get(imgURL)
	if err != nil {
		d.reporter.PrintWarning("ダウンロードエラー: %s - %v", imgURL, err)
		return "", fmt.Errorf("ダウンロードに失敗しました: %v", err)
	}

	if !resp.IsSuccess() {
		// ファイルが作成されていたら削除
		os.Remove(destPath)
		d.reporter.PrintWarning("HTTPエラー: %s - %s", imgURL, resp.Status)
		return "", fmt.Errorf("HTTPエラー: %s", resp.Status)
	}

	d.reporter.PrintInfo("ダウンロード成功: %s", destPath)
	// 相対パスを返す
	return fileName, nil
}

// guessImageExt はURLから画像拡張子を推測
func guessImageExt(url string) string {
	lower := strings.ToLower(url)
	if strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") {
		return ".jpg"
	} else if strings.Contains(lower, ".png") {
		return ".png"
	} else if strings.Contains(lower, ".gif") {
		return ".gif"
	} else if strings.Contains(lower, ".webp") {
		return ".webp"
	} else if strings.Contains(lower, ".svg") {
		return ".svg"
	}
	return ".jpg" // デフォルト
}
