package util

import (
	"fmt"
	"strings"
	"time"
)

// ParseArticleDate は記事の日付文字列を様々なフォーマットでパースする
func ParseArticleDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	dateStr = strings.TrimRight(dateStr, "\\")
	formats := []string{
		"01/02/2006 15:04:05", // 標準的なMT形式 (MM/DD/YYYY)
		"01/02/2006 00:00:00", // 時間が00:00:00の形式
	}
	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return t, fmt.Errorf("could not parse date '%s': %v", dateStr, err)
}

// FormatDirName は日付から年/月/日/時分形式のディレクトリ名を生成する
func FormatDirName(t time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
}

// CreateSlug はタイトルからURLスラグを生成する
func CreateSlug(title string) string {
	slug := strings.ReplaceAll(title, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	slug = strings.ReplaceAll(slug, ":", "-")
	return slug
}
