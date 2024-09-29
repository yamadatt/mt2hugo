package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

// Parse the mobably type export file and extract article data
func parseExportFile(lines []string) []map[string]string {
	var articles []map[string]string
	var article map[string]string

	for _, line := range lines {
		if line == "--------" {
			if article != nil {
				articles = append(articles, article)
			}
			article = make(map[string]string)
		} else {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				article[parts[0]] = parts[1]
				// Handle empty lines
			} else if len(parts) == 1 && parts[0] == "" {
				continue
			}
		}
	}

	if article != nil {
		articles = append(articles, article)
	}

	return articles
}

// Create a directory for each article and write the Hugo file
func createHugoFiles(articles []map[string]string) error {
	for _, article := range articles {
		title, ok := article["TITLE"]
		if !ok {
			return fmt.Errorf("missing TITLE field in article")
		}
		dirName := strings.ReplaceAll(title, " ", "_")
		dirPath := filepath.Join("output", dirName)

		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return err
		}

		filePath := filepath.Join(dirPath, "index.md")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		fmt.Fprintf(writer, "+++\n")
		for key, value := range article {
			fmt.Fprintf(writer, "%s = \"%s\"\n", key, value)
		}
		fmt.Fprintf(writer, "+++\n")
		writer.Flush()
	}

	return nil
}

// Main function to call the above functions and perform the conversion
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path_to_movable_type_export_file>")
		return
	}

	filePath := os.Args[1]
	lines, err := readExportFile(filePath)
	if err != nil {
		fmt.Println("Error reading export file:", err)
		return
	}

	articles := parseExportFile(lines)
	if err := createHugoFiles(articles); err != nil {
		fmt.Println("Error creating Hugo files:", err)
	}
}
