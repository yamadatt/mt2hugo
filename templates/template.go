package templates

import (
	"fmt"
	"os"
	"text/template"
)

// LoadHugoTemplate はHugo用のテンプレートを読み込む
func LoadHugoTemplate(templatePath string) (*template.Template, error) {
	// テンプレートファイルが存在するか確認
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// テンプレートファイルがない場合はエラーを返す
		return nil, fmt.Errorf("テンプレートファイル %s が見つかりません", templatePath)
	}

	// テンプレートファイルを読み込む
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("テンプレートファイルの解析に失敗しました: %w", err)
	}

	return tmpl, nil
}
