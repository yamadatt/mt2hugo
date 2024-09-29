package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"parser"
)

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
	lines, err := parser.readExportFile(filePath)
	if err != nil {
		fmt.Println("Error reading export file:", err)
		return
	}

	articles := parser.parseMovableTypeExportFile(lines)
	if err := createHugoFiles(articles); err != nil {
		fmt.Println("Error creating Hugo files:", err)
	}
}
