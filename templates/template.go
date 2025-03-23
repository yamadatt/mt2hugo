package templates

import (
	"os"
	"text/template"
)

// デフォルトのHugoテンプレート文字列
const DefaultHugoTemplate = `---
title: "{{ .Title }}"
date: {{ .Date }}
slug: "{{ .Slug }}"
{{ if .Category }}category:
  - "{{ .Category }}"
{{ end }}
{{ if .Tags }}tags:
{{ range .Tags }}  - "{{ . }}"
{{ end }}{{ end }}
{{ if .Image }}cover:
    image: "{{ .Image }}"
    alt: "{{ .Title }}"
    hidden: true
    caption: "{{ .Title }}"
{{ end }}
draft: false
showtoc: false
{{ if .Summary }}summary: "{{ .Summary }}"{{ end }}
---

{{ .Body }}

{{ if .ExtendedBody }}
<!--more-->

{{ .ExtendedBody }}
{{ end }}
`

// LoadHugoTemplate はHugo用のテンプレートを読み込む
func LoadHugoTemplate(templatePath string) (*template.Template, error) {
	// テンプレートファイルが存在するか確認
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// テンプレートファイルがない場合はデフォルトテンプレートを使用
		return template.New("hugo").Parse(DefaultHugoTemplate)
	}

	// テンプレートファイルを読み込む
	return template.ParseFiles(templatePath)
}
