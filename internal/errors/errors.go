package errors

import (
	"errors"
	"fmt"
)

// 基本エラー型の定義
var (
	ErrConfiguration  = errors.New("設定エラー")
	ErrFileSystem     = errors.New("ファイルシステムエラー")
	ErrParsing        = errors.New("解析エラー")
	ErrTransform      = errors.New("変換エラー")
	ErrValidation     = errors.New("検証エラー")
	ErrProcessing     = errors.New("処理エラー")
	ErrHTMLConversion = errors.New("HTML変換エラー")
)

// Wrap はエラーをラップする
// err: 元のエラー
// baseErr: ベースとなるエラータイプ
// message: 追加のエラーメッセージ
func Wrap(err error, baseErr error, message string) error {
	return fmt.Errorf("%s: %w: %w", message, baseErr, err)
}

// Is はエラーがターゲットエラーと等しいかをチェックする
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// WrapIf は条件付きでエラーをラップする
func WrapIf(err error, condition bool, baseErr error, message string) error {
	if condition && err != nil {
		return Wrap(err, baseErr, message)
	}
	return err
}

// より汎用的なアプローチ
func WrapIfNotTyped(err error, baseErr error, message string) error {
	// 基本エラータイプの一覧
	baseErrors := []error{
		ErrValidation, ErrConfiguration, ErrParsing,
		ErrTransform, ErrFileSystem, ErrProcessing,
	}

	// 既に基本エラータイプのいずれかでラップされているかチェック
	for _, e := range baseErrors {
		if Is(err, e) {
			return err
		}
	}

	// ラップされていない場合は新しくラップする
	return Wrap(err, baseErr, message)
}

// New は新しいエラーを作成する
func New(message string) error {
	return errors.New(message)
}
