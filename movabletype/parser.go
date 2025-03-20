package movabletype

import (
	"bufio"
	"os"
	"strings"
)

// ReadExportFile はMovable Typeのエクスポートファイルを読み込み、行ごとの配列を返す
func ReadExportFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// ParseExportFile はMovable Typeのエクスポートファイルの行を解析し、
// 各記事のキーと値のマップを含む配列を返す
func ParseExportFile(lines []string) []map[string]string {
	var articles []map[string]string
	var currentArticle map[string]string
	var currentKey string
	var bodyContent string
	var inBody bool
	var inExtendedBody bool // EXTENDED BODY セクションフラグを追加
	var inComment bool      // コメントセクションフラグを追加

	for i, line := range lines {
		// 新しい記事の開始
		if strings.HasPrefix(line, "--------") {
			if currentArticle != nil {
				// BODYがあれば追加
				if bodyContent != "" {
					currentArticle["BODY"] = bodyContent
				}
				articles = append(articles, currentArticle)
			}
			currentArticle = make(map[string]string)
			currentKey = ""
			bodyContent = ""
			inBody = false
			inExtendedBody = false // 新しい記事の初期化時にリセット
			inComment = false      // コメントフラグもリセット
			continue
		}

		// 最初の空行はスキップ
		if currentArticle == nil {
			continue
		}

		// セクションの区切り
		if line == "-----" {
			// COMMENT セクションの開始を検出
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "COMMENT:") {
				inComment = true
				inBody = false
				inExtendedBody = false
				continue
			}

			// BODYセクションの開始を検出
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "BODY:") {
				inBody = true
				inExtendedBody = false
				inComment = false
				continue
			}

			// EXTENDED BODYセクションの開始を検出
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "EXTENDED BODY:") {
				inExtendedBody = true
				inBody = false
				inComment = false
				continue
			}

			// 各セクションの終了
			if inBody || inExtendedBody || inComment {
				inBody = false
				inExtendedBody = false
				inComment = false
				continue
			}
			continue
		}

		// コメントセクションの場合はスキップ
		if inComment {
			continue // コメント行は完全に無視
		}

		// BODYセクションの場合
		if inBody {
			// BODYヘッダーをスキップ
			if strings.HasPrefix(line, "BODY:") {
				continue
			}
			bodyContent += line + "\n"
			continue
		}

		// EXTENDED BODYセクションの場合
		if inExtendedBody {
			// EXTENDED BODYヘッダーをスキップ
			if strings.HasPrefix(line, "EXTENDED BODY:") {
				continue
			}
			// BODY内容と同様に追加
			bodyContent += line + "\n"
			continue
		}

		// キーと値のペアを解析
		if strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				currentKey = parts[0]
				currentArticle[currentKey] = strings.TrimSpace(parts[1])
			}
		} else if currentKey != "" && strings.TrimSpace(line) != "" {
			// 前の値の続き
			currentArticle[currentKey] += " " + strings.TrimSpace(line)
		}
	}

	// 最後の記事を追加
	if currentArticle != nil {
		if bodyContent != "" {
			currentArticle["BODY"] = bodyContent
		}

		// 必須フィールドを持つ記事だけを追加
		if _, hasDate := currentArticle["DATE"]; hasDate {
			articles = append(articles, currentArticle)
		}
	}

	return articles
}
