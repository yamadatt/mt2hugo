package config

import (
	"mt2hugo/transformer/mt2hugo"
)

// Config はアプリケーション全体の設定を保持する構造体
type Config struct {
	// 入力・出力関連
	InputFile    string
	OutputDir    string
	TemplateFile string

	// 変換オプション
	NoMarkdown bool
	FormatHTML bool
	ErrorMode  mt2hugo.ErrorHandlingMode

	// 将来の拡張用
	Verbose bool
}

// NewDefaultConfig はデフォルト設定を持つConfigを返す
func NewDefaultConfig() *Config {
	return &Config{
		OutputDir:    "output",
		TemplateFile: "templates/hugo.tmpl",
		ErrorMode:    mt2hugo.ReturnHTML,
		Verbose:      false,
	}
}

// Validate は設定の整合性を確認する
func (c *Config) Validate() []string {
	var warnings []string

	// formatHTMLは--no-markdownと一緒に使う場合のみ効果がある
	if c.FormatHTML && !c.NoMarkdown {
		warnings = append(warnings,
			"注意: --format-htmlオプションは--no-markdownと一緒に使用した場合のみ効果があります")
	}

	return warnings
}
