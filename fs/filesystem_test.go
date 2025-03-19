package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// テスト用の一時ディレクトリ作成
func setupTestDir(t *testing.T) string {
	tempDir := filepath.Join(os.TempDir(), "mt2hugo_test")
	err := os.RemoveAll(tempDir)
	require.True(t, err == nil || os.IsNotExist(err), "テスト用ディレクトリのクリーンアップに失敗: %v", err)

	err = os.MkdirAll(tempDir, 0755)
	require.NoError(t, err, "テスト用ディレクトリの作成に失敗")

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
	err := os.WriteFile(fullPath, []byte(content), 0644)
	require.NoError(t, err, "テスト用ファイルの作成に失敗")
	return fullPath
}

// FileSystemErrorのテスト
func TestFileSystemError(t *testing.T) {
	testCases := []struct {
		name           string
		fsErr          *FileSystemError
		expectedMsg    string
		checkUnwrap    bool
		expectedUnwrap error
	}{
		{
			name: "元のエラーあり",
			fsErr: &FileSystemError{
				Op:      "TestOp",
				Path:    "/test/path",
				Message: "テストエラーメッセージ",
				Err:     errors.New("元のエラー"),
			},
			expectedMsg:    "ファイルシステムエラー: TestOp /test/path: テストエラーメッセージ - 元のエラー",
			checkUnwrap:    true,
			expectedUnwrap: errors.New("元のエラー"),
		},
		{
			name: "元のエラーなし",
			fsErr: &FileSystemError{
				Op:      "TestOp",
				Path:    "/test/path",
				Message: "テストエラーメッセージ",
				Err:     nil,
			},
			expectedMsg:    "ファイルシステムエラー: TestOp /test/path: テストエラーメッセージ",
			checkUnwrap:    false,
			expectedUnwrap: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Errorメソッドのテスト
			assert.Equal(t, tc.expectedMsg, tc.fsErr.Error(), "Error()メソッドの出力が期待値と一致しません")

			// Unwrapメソッドのテスト（必要な場合）
			if tc.checkUnwrap {
				assert.Equal(t, tc.expectedUnwrap.Error(), tc.fsErr.Unwrap().Error(), "Unwrap()メソッドの出力が期待値と一致しません")
			}
		})
	}
}

// ReadFileのテスト
func TestReadFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	// テスト用ファイル作成
	singleLineFile := createTestFile(t, testDir, "single.txt", "単一行")
	multiLineFile := createTestFile(t, testDir, "multi.txt", "line1\nline2\nline3")
	notExistFile := filepath.Join(testDir, "notexist.txt")

	fs := NewRealFileSystem()

	testCases := []struct {
		name          string
		filePath      string
		expectedLines []string
		expectError   bool
		checkErrorOp  bool
		expectedOp    string
	}{
		{
			name:          "単一行ファイル",
			filePath:      singleLineFile,
			expectedLines: []string{"単一行"},
			expectError:   false,
		},
		{
			name:          "複数行ファイル",
			filePath:      multiLineFile,
			expectedLines: []string{"line1", "line2", "line3"},
			expectError:   false,
		},
		{
			name:          "存在しないファイル",
			filePath:      notExistFile,
			expectedLines: nil,
			expectError:   true,
			checkErrorOp:  true,
			expectedOp:    "ReadFile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := fs.ReadFile(tc.filePath)

			// エラーの有無をチェック
			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")

				// エラー型と操作名をチェック（必要な場合）
				if tc.checkErrorOp {
					var fsErr *FileSystemError
					assert.True(t, errors.As(err, &fsErr), "エラーがFileSystemErrorではありません: %T", err)
					if fsErr != nil {
						assert.Equal(t, tc.expectedOp, fsErr.Op, "エラー操作が期待値と一致しません")
					}
				}
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")

				// 内容をチェック
				assert.Equal(t, tc.expectedLines, lines, "返された行が期待値と一致しません")
			}
		})
	}
}

// WriteFileのテスト
func TestWriteFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()

	testCases := []struct {
		name        string
		filePath    string
		content     string
		expectError bool
		setup       func(*testing.T)                 // セットアップ関数
		validate    func(*testing.T, string, string) // 検証関数
	}{
		{
			name:        "通常のファイル作成",
			filePath:    filepath.Join(testDir, "normal.txt"),
			content:     "テスト内容\n複数行あり\n",
			expectError: false,
			setup:       nil,
			validate: func(t *testing.T, path, expectedContent string) {
				// ファイルが作成されたか確認
				assert.FileExists(t, path, "ファイルが作成されていません")

				// 内容を確認
				data, err := os.ReadFile(path)
				require.NoError(t, err, "ファイルの読み込みに失敗")
				assert.Equal(t, expectedContent, string(data), "ファイル内容が期待値と一致しません")
			},
		},
		{
			name:        "サブディレクトリへの書き込み",
			filePath:    filepath.Join(testDir, "subdir", "file.txt"),
			content:     "サブディレクトリ内のファイル",
			expectError: false,
			setup:       nil,
			validate: func(t *testing.T, path, expectedContent string) {
				// サブディレクトリが作成されたか確認
				dir := filepath.Dir(path)
				assert.DirExists(t, dir, "サブディレクトリが作成されていません")

				// 内容を確認
				data, err := os.ReadFile(path)
				require.NoError(t, err, "ファイルの読み込みに失敗")
				assert.Equal(t, expectedContent, string(data), "ファイル内容が期待値と一致しません")
			},
		},
		{
			name:        "書き込み権限のないディレクトリ",
			filePath:    filepath.Join(testDir, "readonly", "file.txt"),
			content:     "書き込めないはず",
			expectError: os.Geteuid() != 0, // rootユーザー以外の場合のみエラーを期待
			setup: func(t *testing.T) {
				if os.Geteuid() == 0 {
					t.Skip("rootユーザーでは権限エラーをテストできません")
				}
				readonlyDir := filepath.Join(testDir, "readonly")
				err := os.MkdirAll(readonlyDir, 0500)
				require.NoError(t, err, "読み取り専用ディレクトリの作成に失敗")
			},
			validate: nil, // エラーが発生するため検証は不要
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// セットアップ実行（必要な場合）
			if tc.setup != nil {
				tc.setup(t)
			}

			// テスト対象の関数を実行
			err := fs.WriteFile(tc.filePath, tc.content)

			// エラーの有無をチェック
			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")

				// 検証関数を実行（必要な場合）
				if tc.validate != nil {
					tc.validate(t, tc.filePath, tc.content)
				}
			}
		})
	}
}

// TryReadFileのテスト
func TestTryReadFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()

	// テスト用ファイル作成
	existingFile := createTestFile(t, testDir, "existing.txt", "テスト内容")
	notExistFile := filepath.Join(testDir, "notexist.txt")

	testCases := []struct {
		name          string
		filePath      string
		expectedLines []string
		expectError   bool
	}{
		{
			name:          "存在するファイル",
			filePath:      existingFile,
			expectedLines: []string{"テスト内容"},
			expectError:   false,
		},
		{
			name:          "存在しないファイル",
			filePath:      notExistFile,
			expectedLines: []string{},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// テスト対象の関数を実行
			lines, err := fs.TryReadFile(tc.filePath)

			// エラーの有無をチェック
			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")
			}

			// 返された行のチェック
			assert.Equal(t, tc.expectedLines, lines, "返された行が期待値と一致しません")
		})
	}
}

// TryWriteFileのテスト
func TestTryWriteFile(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()

	testCases := []struct {
		name        string
		filePath    string
		content     string
		setup       func(*testing.T)
		validate    func(*testing.T, string, string)
		expectError bool
	}{
		{
			name:        "通常のファイル作成",
			filePath:    filepath.Join(testDir, "normal.txt"),
			content:     "テスト内容",
			expectError: false,
			validate: func(t *testing.T, path, expectedContent string) {
				data, err := os.ReadFile(path)
				require.NoError(t, err, "ファイルの読み込みに失敗")
				assert.Equal(t, expectedContent, string(data), "ファイル内容が期待値と一致しません")
			},
		},
		{
			name:        "書き込み権限のないディレクトリ",
			filePath:    filepath.Join(testDir, "readonly", "file.txt"),
			content:     "書き込めないはず",
			expectError: os.Geteuid() != 0,
			setup: func(t *testing.T) {
				if os.Geteuid() == 0 {
					t.Skip("rootユーザーでは権限エラーをテストできません")
				}
				readonlyDir := filepath.Join(testDir, "readonly")
				err := os.MkdirAll(readonlyDir, 0500)
				require.NoError(t, err, "読み取り専用ディレクトリの作成に失敗")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// セットアップ実行（必要な場合）
			if tc.setup != nil {
				tc.setup(t)
			}

			// テスト対象の関数を実行
			err := fs.TryWriteFile(tc.filePath, tc.content)

			// エラーの有無をチェック
			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")

				// 検証関数を実行（必要な場合）
				if tc.validate != nil {
					tc.validate(t, tc.filePath, tc.content)
				}
			}
		})
	}
}

// MkdirAllのテスト
func TestMkdirAll(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	fs := NewRealFileSystem()

	testCases := []struct {
		name        string
		dirPath     string
		setup       func(*testing.T)
		validate    func(*testing.T, string)
		expectError bool
	}{
		{
			name:        "通常のディレクトリ作成",
			dirPath:     filepath.Join(testDir, "a", "b", "c"),
			expectError: false,
			validate: func(t *testing.T, path string) {
				assert.DirExists(t, path, "ディレクトリが作成されていません")
			},
		},
		{
			name:    "既存のディレクトリ",
			dirPath: filepath.Join(testDir, "existing"),
			setup: func(t *testing.T) {
				err := os.MkdirAll(filepath.Join(testDir, "existing"), 0755)
				require.NoError(t, err, "既存ディレクトリの作成に失敗")
			},
			expectError: false,
			validate: func(t *testing.T, path string) {
				assert.DirExists(t, path, "ディレクトリが存在しません")
			},
		},
		{
			name:        "権限のないパス",
			dirPath:     "/root/test-dir",
			expectError: os.Geteuid() != 0 && fileExists("/root"),
			setup: func(t *testing.T) {
				if os.Geteuid() == 0 {
					t.Skip("rootユーザーでは権限エラーをテストできません")
				}
				if !fileExists("/root") {
					t.Skip("rootディレクトリが存在しないためテストをスキップします")
				}
			},
			validate: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// セットアップ実行（必要な場合）
			if tc.setup != nil {
				tc.setup(t)
			}

			// テスト対象の関数を実行
			err := fs.MkdirAll(tc.dirPath)

			// エラーの有無をチェック
			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")

				// エラーがFileSystemErrorか確認
				var fsErr *FileSystemError
				assert.True(t, errors.As(err, &fsErr), "エラーがFileSystemErrorではありません")
				if fsErr != nil {
					assert.Equal(t, "MkdirAll", fsErr.Op, "エラー操作が期待値と一致しません")
				}
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")

				// 検証関数を実行（必要な場合）
				if tc.validate != nil {
					tc.validate(t, tc.dirPath)
				}
			}
		})
	}
}

// ファイルやディレクトリが存在するか確認するヘルパー関数
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// NewRealFileSystemのテスト
func TestNewRealFileSystem(t *testing.T) {
	t.Run("正しい型を返すこと", func(t *testing.T) {
		fs := NewRealFileSystem()
		_, ok := interface{}(fs).(*RealFileSystem)
		assert.True(t, ok, "NewRealFileSystemが正しい型を返しませんでした: %T", fs)
	})

	t.Run("インターフェースを満たすこと", func(t *testing.T) {
		fs := NewRealFileSystem()
		var fsInterface FileSystem = fs
		assert.NotNil(t, fsInterface, "RealFileSystemがFileSystemインターフェースを満たしていません")
	})
}

// モックファイルシステムテスト
func TestMockFileSystem(t *testing.T) {
	// モックファイルシステムがインターフェースを満たしていることを確認
	t.Run("インターフェースを満たすこと", func(t *testing.T) {
		var _ FileSystem = &MockFileSystem{}
	})

	testCases := []struct {
		name           string
		mockFunc       func() *MockFileSystem
		operation      func(*MockFileSystem) (interface{}, error)
		expectedResult interface{}
		expectError    bool
	}{
		{
			name: "ReadFile - 成功",
			mockFunc: func() *MockFileSystem {
				return &MockFileSystem{
					ReadFileFunc: func(path string) ([]string, error) {
						return []string{"mocked line"}, nil
					},
				}
			},
			operation: func(m *MockFileSystem) (interface{}, error) {
				return m.ReadFile("dummy")
			},
			expectedResult: []string{"mocked line"},
			expectError:    false,
		},
		{
			name: "ReadFile - エラー",
			mockFunc: func() *MockFileSystem {
				return &MockFileSystem{
					ReadFileFunc: func(path string) ([]string, error) {
						return nil, errors.New("mock error")
					},
				}
			},
			operation: func(m *MockFileSystem) (interface{}, error) {
				return m.ReadFile("dummy")
			},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name: "WriteFile - 成功",
			mockFunc: func() *MockFileSystem {
				return &MockFileSystem{
					WriteFileFunc: func(path string, content string) error {
						return nil
					},
				}
			},
			operation: func(m *MockFileSystem) (interface{}, error) {
				return nil, m.WriteFile("dummy", "test")
			},
			expectedResult: nil,
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := tc.mockFunc()
			result, err := tc.operation(mock)

			if tc.expectError {
				assert.Error(t, err, "エラーが発生しませんでした。エラーを期待していました")
			} else {
				assert.NoError(t, err, "予期しないエラーが発生しました")
				assert.Equal(t, tc.expectedResult, result, "結果が期待値と一致しません")
			}
		})
	}
}

// MockFileSystem の実装
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
