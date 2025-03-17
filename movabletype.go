package main

import (
	"bufio"
	"os"
	"strings"
)

// Movable Typeのエクスポートファイルを読み込み、行ごとの配列を返す
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

// Movable Typeのエクスポートファイルの行を解析し、
// 各記事のキーと値のマップを含む配列を返す
func ParseMovableTypeExportFile(lines []string) []map[string]string {
	var articles []map[string]string
	var currentArticle map[string]string
	var currentKey string
	var bodyContent string
	var inBody bool

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
			continue
		}

		// 最初の空行はスキップ
		if currentArticle == nil {
			continue
		}

		// セクションの区切り
		if line == "-----" {
			// BODYセクションの開始を検出
			if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "BODY:") {
				inBody = true
				continue
			}
			// BODYセクションの終了
			if inBody {
				inBody = false
				continue
			}
			continue
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
		articles = append(articles, currentArticle)
	}

	return articles
}
