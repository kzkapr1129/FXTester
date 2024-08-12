package internal

import (
	"errors"
	"fmt"
	"fxtester/internal/gen"
	"net/http"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

// エラーコード
type ErrorCode uint32

// エラータイプ
type ErrorType uint8

// エラー別詳細設定
type errorSettings struct {
	statusCode                int
	dictKey                   string
	isDisplayErrCodeOnMessage bool
}

/**
  エラーコードの構成について:

  0x8abbcccc の形式でエラーコードを構成します。

  - a: エラーの原因 (1桁目)
    - 1: サーバー起因のエラー
    - 0: クライアント起因のエラー

  - b: エラーメッセージのタイプ (2桁目)
    - 00: インターナルエラー
    - その他の値は、随時追加されるタイプ

  - c: エラーの詳細番号 (3桁目以降)
*/

const (
	// サーバー起因のエラー
	ErrCodePanic              ErrorCode = 0x80000001
	ErrCodeUnknownErrorObject ErrorCode = 0x80000002 // 不明なエラーオブジェクト
	ErrCodeUnknownErrorCode   ErrorCode = 0x80000003 // 不明なエラーコード
	ErrCodeConfig             ErrorCode = 0x80000004 // 設定ファイル不備

	// ユーザ起因のエラー
	ErrCodeForbiddenCharacterError ErrorCode = 0x81010001 // 禁止文字エラー
)

const (
	DictKeyInternalError           = "InternalError"           // 辞書キー: インターナルエラー
	DictKeyForbiddenCharacterError = "ForbiddenCharacterError" // 辞書キー: 禁止文字エラー
)

var errSettingsMap = map[ErrorType]errorSettings{
	0x00: {
		statusCode:                http.StatusInternalServerError,
		dictKey:                   DictKeyInternalError,
		isDisplayErrCodeOnMessage: true,
	},
	0x01: {
		statusCode:                http.StatusBadRequest,
		dictKey:                   DictKeyForbiddenCharacterError,
		isDisplayErrCodeOnMessage: true,
	},
}

type FxtError struct {
	ErrCode   ErrorCode
	Arguments []interface{}
	cause     error
}

func NewFxtError(errorCode ErrorCode, arguments ...interface{}) *FxtError {
	return &FxtError{
		ErrCode:   errorCode,
		Arguments: arguments,
	}
}

func (e *FxtError) SetCause(err error) *FxtError {
	e.cause = err
	return e
}

func (e *FxtError) Cause() error {
	return e.cause
}

func (e *FxtError) Unwrap() error {
	return e.cause
}

func (e *FxtError) Error() string {
	arguments := strings.Join(ArrayMap(func(input interface{}) string {
		return fmt.Sprintf("\"%v\"", input)
	}, e.Arguments), ",")
	return fmt.Sprintf("{\"code\": \"0x%x\", \"arguments\": [%s]}", e.ErrCode, arguments)
}

// CauseFxtError 原因となったFxtErrorを返却します
func CauseFxtError(err error) *FxtError {
	type causer interface {
		Cause() error
	}

	var lastFxtError *FxtError = nil
	errors.As(err, &lastFxtError)

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
		if err != nil {
			var fxtError *FxtError = nil
			if errors.As(err, &fxtError) {
				lastFxtError = fxtError
			}
		}
	}
	return lastFxtError
}

// ErrorHandler Error型エラーがEchoハンドラーから返却された時、エラーメッセージを多言語化しクライアントに返却する
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {
			defer func() {
				if err := recover(); err != nil {
					// スタックトレースのログ出力
					logStackTrace(c)
					// エラーレスポンス返却
					statusCode, errObject := makeErrorResponseWithCode(c, ErrCodePanic, []interface{}{})
					returnErr = c.JSON(statusCode, errObject)
				}
			}()

			err := next(c)
			if err != nil {
				// エラーレスポンス返却
				statusCode, errObject := makeErrorResponse(c, err)
				return c.JSON(statusCode, errObject)
			}
			return nil
		}
	}
}

func logStackTrace(ctx echo.Context) {
	stack := make([]byte, 4*1024)
	length := runtime.Stack(stack, true)
	stack = stack[:length]
	ctx.Echo().Logger.Error(string(stack))
}

func extractErrorType(errCode ErrorCode) ErrorType {
	return ErrorType(((errCode & 0x00FF0000) >> 16) & 0xFF)
}

func isValidErrorCode(errCode ErrorCode) bool {
	head := errCode & 0x8F000000
	if head != 0x81000000 && head != 0x80000000 {
		return false
	}
	if _, ok := errSettingsMap[extractErrorType(errCode)]; !ok {
		return false
	}
	detailCode := errCode & 0x0000FFFF
	return detailCode != 0x00000000
}

func makeErrorResponseWithCode(ctx echo.Context, errCode ErrorCode, arguments []interface{}) (int, *gen.Error) {
	if !isValidErrorCode(errCode) {
		panic(fmt.Sprintf("invalid error code at makeErrorResponseWithCode: 0x%x", errCode))
	}
	settings := errSettingsMap[extractErrorType(errCode)]

	// 不明なエラーオブジェクトの場合、インターナルエラーとして扱う
	return settings.statusCode, &gen.Error{
		Code: uint32(errCode),
		Message: GetDict(ctx,
			[]string{
				"messages",
				settings.dictKey,
			},
			func() []interface{} {
				if settings.isDisplayErrCodeOnMessage {
					arguments = append(arguments, errCode)
				}
				return arguments
			}()...),
	}
}

func makeErrorResponse(ctx echo.Context, err error) (int, *gen.Error) {
	fxtError := CauseFxtError(err)
	if fxtError == nil {
		// 不明なエラーオブジェクトの場合
		ctx.Echo().Logger.Errorf("caught invalid Error: %v", err)
		return makeErrorResponseWithCode(ctx, ErrCodeUnknownErrorObject, []interface{}{})
	}

	errCode := fxtError.ErrCode
	arguments := fxtError.Arguments
	settings, ok := errSettingsMap[extractErrorType(fxtError.ErrCode)]
	if !ok || !isValidErrorCode(errCode) {
		// 未登録のエラーコードを受け取った場合
		ctx.Echo().Logger.Errorf("caught invalid Error: %v", fxtError)

		//　エラー内容の置き換え
		settings = errorSettings{
			statusCode:                http.StatusInternalServerError,
			dictKey:                   DictKeyInternalError,
			isDisplayErrCodeOnMessage: true,
		}
		errCode = ErrCodeUnknownErrorCode
		arguments = []interface{}{}
	}

	if settings.isDisplayErrCodeOnMessage {
		arguments = append(arguments, errCode)
	}

	// gen.Errorオブジェクトに多言語化対応済みメッセージを格納し返却する
	return settings.statusCode, &gen.Error{
		Code: uint32(errCode),
		Message: GetDict(ctx, []string{
			"messages",
			settings.dictKey}, GetDicts(ctx, arguments)...),
	}
}
