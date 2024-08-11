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

// エラー別詳細設定
type errorSettings struct {
	statusCode                int
	dictKey                   string
	isDisplayErrCodeOnMessage bool
}

const (
	ErrCodePanic                   ErrorCode = 0x80000000
	ErrCodeUnknownErrorObject      ErrorCode = 0x80000001 // 不明なエラーオブジェクト
	ErrCodeUnknownErrorCode        ErrorCode = 0x80000002 // 不明なエラーコード
	ErrCodeForbiddenCharacterError ErrorCode = 0x81000003 // 禁止文字エラー
)

const (
	DictKeyInternalError           = "InternalError"           // 辞書キー: インターナルエラー
	DictKeyForbiddenCharacterError = "ForbiddenCharacterError" // 辞書キー: 禁止文字エラー
)

var errSettingsMap = map[ErrorCode]errorSettings{
	ErrCodePanic: {
		statusCode:                http.StatusInternalServerError,
		dictKey:                   DictKeyInternalError,
		isDisplayErrCodeOnMessage: true,
	},
	ErrCodeUnknownErrorObject: {
		statusCode:                http.StatusInternalServerError,
		dictKey:                   DictKeyInternalError,
		isDisplayErrCodeOnMessage: true,
	},
	ErrCodeUnknownErrorCode: {
		statusCode:                http.StatusInternalServerError,
		dictKey:                   DictKeyInternalError,
		isDisplayErrCodeOnMessage: true,
	},
	ErrCodeForbiddenCharacterError: {
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

func makeErrorResponseWithCode(ctx echo.Context, errCode ErrorCode, arguments []interface{}) (int, *gen.Error) {
	settings := errSettingsMap[errCode]

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
	settings, ok := errSettingsMap[fxtError.ErrCode]
	if !ok {
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
