package internal

import (
	"strings"

	"github.com/labstack/echo/v4"
)

type Error struct {
	message string
	cause   error
}

func NewError(ctx echo.Context, msgType string, arguments ...interface{}) *Error {
	// パラメータに辞書キーが指定されている場合は変換する
	newArguments := ArrayMap(func(input interface{}) interface{} {
		if keys, ok := input.(string); ok && strings.Contains(keys, ".") {
			return GetDict(ctx, strings.Split(keys, "."))
		}
		return input
	}, arguments)

	// エラーメッセージの生成
	message := GetDict(ctx, []string{
		"messages",
		msgType,
	}, newArguments...)

	return &Error{
		message: message,
	}
}

func (e *Error) SetCause(err error) *Error {
	e.cause = err
	return e
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Error() string {
	return e.message
}
