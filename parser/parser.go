package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Read the mobably type export file
func readExportFile(filePath string) ([]string, error) {
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

// Parse the MovableType export file and extract article data
func parseMovableTypeExportFile(lines []string) []map[string]string {
	var articles []map[string]string
	var article map[string]string
	var bodyContent []string
	inBody := false

	for _, line := range lines {
		if line == "--------" {
			if article != nil {
				article["BODY"] = strings.Join(bodyContent, "\n")
				articles = append(articles, article)
			}
			article = make(map[string]string)
			bodyContent = nil
			inBody = false
		} else {
			if inBody {
				bodyContent = append(bodyContent, line)
			} else {
				parts := strings.SplitN(line, ": ", 2)
				if len(parts) == 2 {
					if article == nil {
						article = make(map[string]string)
					}
					article[parts[0]] = parts[1]
					if parts[0] == "BODY" {
						inBody = true
					}
				} else if len(parts) == 1 && parts[0] == "" {
					continue
				}
			}
		}
	}

	if article != nil {
		article["BODY"] = strings.Join(bodyContent, "\n")
		articles = append(articles, article)
	}

	// Output the specified debug items to standard output
	for _, article := range articles {
		fmt.Println("AUTHOR:", article["AUTHOR"])
		fmt.Println("TITLE:", article["TITLE"])
		fmt.Println("BASENAME:", article["BASENAME"])
		fmt.Println("STATUS:", article["STATUS"])
		fmt.Println("ALLOW COMMENTS:", article["ALLOW COMMENTS"])
		fmt.Println("CONVERT BREAKS:", article["CONVERT BREAKS"])
		fmt.Println("DATE:", article["DATE"])
		fmt.Println("CATEGORY:", article["CATEGORY"])
		fmt.Println("IMAGE:", article["IMAGE"])
		fmt.Println("BODY:", article["BODY"])
		fmt.Println("--------")
	}

	return articles
}
