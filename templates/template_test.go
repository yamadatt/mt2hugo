package templates

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadHugoTemplate(t *testing.T) {
	// テスト用のディレクトリを作成
	tempDir := t.TempDir()

	// 正常なテンプレートファイル
	validTemplatePath := filepath.Join(tempDir, "valid_template.tmpl")
	validTemplateContent := `Hello {{ .Name }}!`
	require.NoError(t, os.WriteFile(validTemplatePath, []byte(validTemplateContent), 0644))

	// 不正なテンプレートファイル（構文エラーあり）
	invalidTemplatePath := filepath.Join(tempDir, "invalid_template.tmpl")
	invalidTemplateContent := `Hello {{ .Name !` // 閉じ括弧が欠けている
	require.NoError(t, os.WriteFile(invalidTemplatePath, []byte(invalidTemplateContent), 0644))

	// 存在しないファイルパス
	nonExistentPath := filepath.Join(tempDir, "non_existent_template.tmpl")

	tests := []struct {
		name         string
		templatePath string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "正常なテンプレートの読み込み",
			templatePath: validTemplatePath,
			wantErr:      false,
		},
		{
			name:         "存在しないテンプレートファイル",
			templatePath: nonExistentPath,
			wantErr:      true,
			errContains:  "が見つかりません",
		},
		{
			name:         "不正なテンプレートファイル",
			templatePath: invalidTemplatePath,
			wantErr:      true,
			errContains:  "テンプレートファイルの解析に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := LoadHugoTemplate(tt.templatePath)

			if tt.wantErr {
				assert.Error(t, err, "エラーが発生するべき")
				assert.Contains(t, err.Error(), tt.errContains, "エラーメッセージに期待した文字列が含まれていること")
				assert.Nil(t, tmpl, "エラー時はテンプレートがnilであること")
			} else {
				assert.NoError(t, err, "エラーが発生しないこと")
				assert.NotNil(t, tmpl, "テンプレートがnilでないこと")

				// 正常に読み込めた場合は簡単な実行テスト
				var data = struct{ Name string }{"World"}
				var output string
				buffer := new(bytes.Buffer)
				err = tmpl.Execute(buffer, data)
				require.NoError(t, err, "テンプレートの実行に失敗しないこと")
				output = buffer.String()
				assert.Equal(t, "Hello World!", output, "テンプレートが正しく実行されること")
			}
		})
	}
}
