package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// テスト用の一時ディレクトリ作成
func setupTestDir(t *testing.T) string {
	tempDir := filepath.Join(os.TempDir(), "mt2hugo_test")
	if err := os.RemoveAll(tempDir); err != nil && !os.IsNotExist(err) {
		t.Fatalf("テスト用ディレクトリのクリーンアップに失敗: %v", err)
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("テスト用ディレクトリの作成に失敗: %v", err)
	}
	return tempDir
}

// テスト後の片付け
func cleanupTestDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("テスト用ディレクトリの削除に失敗: %v", err)
	}
}

// テスト用のファイル作成
func createTestFile(t *testing.T, dir, filename, content string) string {
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("テスト用ファイルの作成に失敗: %v", err)
	}
	return fullPath
}

// FileSystemErrorのテスト
func TestFileSystemError(t *testing.T) {
	origErr := errors.New("元のエラー")
	fsErr := &FileSystemError{
		Op:      "TestOp",
		Path:    "/test/path",
		Message: "テストエラーメッセージ",
		Err:     origErr,
	}

	// Errorメソッドのテスト
	expected := "ファイルシステムエラー: TestOp /test/path: テストエラーメッセージ - 元のエラー"
	if fsErr.Error() != expected {
		t.Errorf("Error()が期待値と一致しません。\n期待値: %s\n実際値: %s", expected, fsErr.Error())
	}

	// Unwrapメソッドのテスト
	if fsErr.Unwrap() != origErr {
		t.Errorf("Unwrap()が元のエラーを返しませんでした")
	}

	// Errが無い場合
	fsErr.Err = nil
	expectedNoErr := "ファイルシステムエラー: TestOp /test/path: テストエラーメッセージ"
	if fsErr.Error() != expectedNoErr {
		t.Errorf("Error()が期待値と一致しません(Errなし)。\n期待値: %s\n実際値: %s", expectedNoErr, fsErr.Error())
	}
}

// ReadFileのテスト
func TestReadFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	// テスト用ファイル作成
	content := "line1\nline2\nline3"
	testFile := createTestFile(t, testDir, "test.txt", content)

	fs := NewRealFileSystem()

	// 正常ケース
	lines, err := fs.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFileが失敗: %v", err)
	}

	expected := []string{"line1", "line2", "line3"}
	if len(lines) != len(expected) {
		t.Fatalf("行数が期待値と一致しません: 期待値=%d, 実際値=%d", len(expected), len(lines))
	}

	for i, line := range expected {
		if lines[i] != line {
			t.Errorf("行 %d が期待値と一致しません: 期待値=%s, 実際値=%s", i, line, lines[i])
		}
	}

	// エラーケース - 存在しないファイル
	_, err = fs.ReadFile(filepath.Join(testDir, "notexist.txt"))
	if err == nil {
		t.Error("存在しないファイルでエラーが発生しませんでした")
	}

	// エラー型の検証
	var fsErr *FileSystemError
	if !errors.As(err, &fsErr) {
		t.Errorf("エラーがFileSystemErrorではありません: %T", err)
	} else {
		if fsErr.Op != "ReadFile" {
			t.Errorf("エラー操作が期待値と一致しません: 期待値=ReadFile, 実際値=%s", fsErr.Op)
		}
	}
}

// WriteFileのテスト
func TestWriteFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()
	testFile := filepath.Join(testDir, "output.txt")
	content := "テスト内容\n複数行あり\n"

	// 正常ケース
	err := fs.WriteFile(testFile, content)
	if err != nil {
		t.Fatalf("WriteFileが失敗: %v", err)
	}

	// ファイルが作成されたことを確認
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("作成されたファイルの読み込みに失敗: %v", err)
	}

	if string(data) != content {
		t.Errorf("ファイル内容が期待値と一致しません:\n期待値=%q\n実際値=%q", content, string(data))
	}

	// サブディレクトリへの書き込みテスト
	testSubFile := filepath.Join(testDir, "subdir", "output.txt")
	err = fs.WriteFile(testSubFile, content)
	if err != nil {
		t.Fatalf("サブディレクトリへのWriteFileが失敗: %v", err)
	}

	// サブディレクトリが作成されたことを確認
	if _, err := os.Stat(filepath.Join(testDir, "subdir")); err != nil {
		t.Errorf("サブディレクトリが作成されませんでした: %v", err)
	}

	// エラーケース - 書き込み権限のないディレクトリ
	if os.Geteuid() == 0 { // rootユーザーの場合はスキップ
		t.Skip("rootユーザーでは権限エラーをテストできません")
	}

	// 書き込み権限のないディレクトリを作成
	readonlyDir := filepath.Join(testDir, "readonly")
	if err := os.MkdirAll(readonlyDir, 0500); err != nil {
		t.Fatalf("読み取り専用ディレクトリの作成に失敗: %v", err)
	}

	readonlyFile := filepath.Join(readonlyDir, "test.txt")
	err = fs.WriteFile(readonlyFile, content)
	if err == nil {
		// 権限の問題でエラーが発生するはず
		t.Error("書き込み権限のないパスでエラーが発生しませんでした")
	}
}

// TryReadFileのテスト
func TestTryReadFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	// テスト用ファイル作成
	content := "テスト内容"
	testFile := createTestFile(t, testDir, "test.txt", content)

	fs := NewRealFileSystem()

	// 正常ケース
	lines, err := fs.TryReadFile(testFile)
	if err != nil {
		t.Fatalf("TryReadFileが失敗: %v", err)
	}

	if len(lines) != 1 || lines[0] != content {
		t.Errorf("ファイル内容が期待値と一致しません: 期待値=[%s], 実際値=%v", content, lines)
	}

	// 存在しないファイル
	lines, err = fs.TryReadFile(filepath.Join(testDir, "notexist.txt"))
	if err == nil {
		t.Error("存在しないファイルでエラーが発生しませんでした")
	}
	if len(lines) != 0 {
		t.Errorf("エラー時に空のスライスが返されませんでした: %v", lines)
	}
}

// TryWriteFileのテスト
func TestTryWriteFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()
	testFile := filepath.Join(testDir, "output.txt")
	content := "テスト内容"

	// 正常ケース
	err := fs.TryWriteFile(testFile, content)
	if err != nil {
		t.Fatalf("TryWriteFileが失敗: %v", err)
	}

	// ファイル内容の確認
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("作成されたファイルの読み込みに失敗: %v", err)
	}
	if string(data) != content {
		t.Errorf("ファイル内容が期待値と一致しません: 期待値=%s, 実際値=%s", content, string(data))
	}

	// エラーケースでも関数は完了する
	if os.Geteuid() != 0 { // rootユーザー以外の場合
		readonlyDir := filepath.Join(testDir, "readonly")
		if err := os.MkdirAll(readonlyDir, 0500); err != nil {
			t.Fatalf("読み取り専用ディレクトリの作成に失敗: %v", err)
		}

		readonlyFile := filepath.Join(readonlyDir, "test.txt")
		err = fs.TryWriteFile(readonlyFile, content)
		// エラーは返すが、関数は正常に戻る
		if err == nil {
			t.Error("書き込み権限のないパスでエラーが発生しませんでした")
		}
	}
}

// MkdirAllのテスト
func TestMkdirAll(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()
	testSubDir := filepath.Join(testDir, "a", "b", "c")

	// 正常ケース
	err := fs.MkdirAll(testSubDir)
	if err != nil {
		t.Fatalf("MkdirAllが失敗: %v", err)
	}

	// ディレクトリが作成されたことを確認
	if _, err := os.Stat(testSubDir); err != nil {
		t.Errorf("ディレクトリが作成されませんでした: %v", err)
	}

	// 既存のディレクトリに対して実行
	err = fs.MkdirAll(testSubDir)
	if err != nil {
		t.Errorf("既存のディレクトリに対するMkdirAllがエラーを返しました: %v", err)
	}

	// エラーケース - 権限のないパス
	if os.Geteuid() == 0 {
		t.Skip("rootユーザーでは権限エラーをテストできません")
	}

	// Unixシステムでのみ有効なテスト
	if _, err := os.Stat("/root"); err == nil {
		err = fs.MkdirAll("/root/test-dir")
		if err == nil {
			t.Error("権限のないパスでエラーが発生しませんでした")
		}

		var fsErr *FileSystemError
		if !errors.As(err, &fsErr) {
			t.Errorf("エラーがFileSystemErrorではありません: %T", err)
		} else {
			if fsErr.Op != "MkdirAll" {
				t.Errorf("エラー操作が期待値と一致しません: 期待値=MkdirAll, 実際値=%s", fsErr.Op)
			}
		}
	}
}

// NewRealFileSystemのテスト
func TestNewRealFileSystem(t *testing.T) {
	fs := NewRealFileSystem()

	// 正しい型を返すことを確認
	_, ok := fs.(*RealFileSystem)
	if !ok {
		t.Errorf("NewRealFileSystemが正しい型を返しませんでした: %T", fs)
	}

	// インターフェースを満たしていることを確認
	var _ FileSystem = fs
}

// モックファイルシステムテスト（オプション）
type MockFileSystem struct {
	ReadFileFunc     func(path string) ([]string, error)
	WriteFileFunc    func(path string, content string) error
	MkdirAllFunc     func(path string) error
	TryReadFileFunc  func(path string) ([]string, error)
	TryWriteFileFunc func(path string, content string) error
}

func (m *MockFileSystem) ReadFile(path string) ([]string, error) {
	return m.ReadFileFunc(path)
}

func (m *MockFileSystem) WriteFile(path string, content string) error {
	return m.WriteFileFunc(path, content)
}

func (m *MockFileSystem) MkdirAll(path string) error {
	return m.MkdirAllFunc(path)
}

func (m *MockFileSystem) TryReadFile(path string) ([]string, error) {
	return m.TryReadFileFunc(path)
}

func (m *MockFileSystem) TryWriteFile(path string, content string) error {
	return m.TryWriteFileFunc(path, content)
}

func TestMockFileSystem(t *testing.T) {
	// モックファイルシステムがインターフェースを満たしていることを確認
	var _ FileSystem = &MockFileSystem{}

	mock := &MockFileSystem{
		ReadFileFunc: func(path string) ([]string, error) {
			return []string{"mocked line"}, nil
		},
	}

	lines, err := mock.ReadFile("dummy")
	if err != nil {
		t.Errorf("モックReadFileが失敗: %v", err)
	}
	if len(lines) != 1 || lines[0] != "mocked line" {
		t.Errorf("モックが期待値を返しませんでした: %v", lines)
	}
}
