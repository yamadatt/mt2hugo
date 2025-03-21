package hugo

import (
	"fmt"

	"mt2hugo/converter"
)

// ContentProcessor はHTMLコンテンツの処理を担当する
type ContentProcessor struct {
	converter  converter.HTMLConverter
	noMarkdown bool
	formatHTML bool
}

// NewContentProcessor は新しいContentProcessorを作成する
func NewContentProcessor(htmlConverter converter.HTMLConverter, noMarkdown bool, formatHTML bool) *ContentProcessor {
	return &ContentProcessor{
		converter:  htmlConverter,
		noMarkdown: noMarkdown,
		formatHTML: formatHTML,
	}
}

// ProcessContent は記事の本文と拡張本文を処理する
func (p *ContentProcessor) ProcessContent(body string, extendedBody string) (string, string, error) {
	// 本文の処理
	processedBody, err := p.processHtmlContent(body)
	if err != nil {
		return "", "", fmt.Errorf("本文処理エラー: %v", err)
	}

	// 拡張本文の処理
	var processedExtendedBody string
	if extendedBody != "" {
		processedExtendedBody, err = p.processHtmlContent(extendedBody)
		if err != nil {
			return "", "", fmt.Errorf("拡張本文処理エラー: %v", err)
		}
	}

	return processedBody, processedExtendedBody, nil
}

// HTMLコンテンツを処理する内部メソッド
func (p *ContentProcessor) processHtmlContent(content string) (string, error) {
	if p.noMarkdown {
		// HTMLをそのまま出力（整形オプションの有無で処理が変わる）
		if p.formatHTML {
			formatted, err := p.converter.FormatHTMLWithIndentation(content)
			if err != nil {
				return "", fmt.Errorf("HTML整形エラー: %v", err)
			}
			return formatted, nil
		}
		return content, nil
	}

	// HTMLをMarkdownに変換
	markdown, err := p.converter.ConvertHTMLToMarkdown(content)
	if err != nil {
		return "", fmt.Errorf("Markdown変換エラー: %v", err)
	}
	return markdown, nil
}
