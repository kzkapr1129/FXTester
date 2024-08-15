package lang

import (
	"errors"
	"fmt"
	"fxtester/internal/common"
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
	ErrCodePanic                  ErrorCode = 0x80000001
	ErrCodeUnknownErrorObject     ErrorCode = 0x80000002 // 不明なエラーオブジェクト
	ErrCodeUnknownErrorCode       ErrorCode = 0x80000003 // 不明なエラーコード
	ErrCodeConfig                 ErrorCode = 0x80000004 // 設定ファイル不備
	ErrCodeDisk                   ErrorCode = 0x80000005 // ファイル読み込みエラー
	ErrInvalidIdpMetadata         ErrorCode = 0x80000006 // idPのメタデータの解析に失敗した場合
	ErrDownloadIdpMetadata        ErrorCode = 0x80000007 // idpのメタデータのダウンロードに失敗した場合
	ErrSSOAuthnRequest            ErrorCode = 0x80000008 // SSOのAuthnRequest作成に失敗した場合
	ErrSSOHtmlWriting             ErrorCode = 0x80000009 // SSOのHTML書き込み中のエラー
	ErrCookieNone                 ErrorCode = 0x80000010 // クッキーが見つからなかった場合のエラー
	ErrCodeSSOParseResponse       ErrorCode = 0x80000011 // SAMLレスポンスのパースに失敗した場合のエラー
	ErrRequestParse               ErrorCode = 0x80000012 // parseFormの呼び出しエラー
	ErrUnexpectedAssertion        ErrorCode = 0x80000013 // 予期しないSAMLアサーションを取得した場合のエラー
	ErrDBOpen                     ErrorCode = 0x80000014 // DBのOpenエラー
	ErrDBBegin                    ErrorCode = 0x80000015 // DBのトランザクション開始エラー
	ErrDBRollback                 ErrorCode = 0x80000016 // DBのロールバック失敗
	ErrDBCommit                   ErrorCode = 0x80000017 // DBのコミット失敗
	ErrDBQuery                    ErrorCode = 0x80000018 // DBのクエリエラー
	ErrDBQueryResult              ErrorCode = 0x80000019 // DBのクエリ結果のエラー
	ErrSession                    ErrorCode = 0x80000020 // 不正なセッションエラー (TODO クライアントエラー化)
	ErrSLOAuthnRequest            ErrorCode = 0x80000021 // SLOのAuthnRequest作成に失敗した場合
	ErrSLOValidation              ErrorCode = 0x80000022 // SLOのSAMLResponseのバリデーションに失敗した場合
	ErrJWTSign                    ErrorCode = 0x80000023 // JWTのSignに失敗した場合
	ErrBase64SamlRequest          ErrorCode = 0x80000024 // SAMLリクエストのbase64デコード失敗
	ErrBase64SamlResponse         ErrorCode = 0x80000025 // SAMLレスポンスのbase64デコード失敗
	ErrUnmarshalSamlRequest       ErrorCode = 0x80000026 // SAMLリクエストのUnmarshal失敗
	ErrUnmarshalSamlResponse      ErrorCode = 0x80000027 // SAMLリクエストのUnmarshal失敗
	ErrSamlLogoutResponseCreation ErrorCode = 0x80000028 // SAMLログアウトレスポンスの作成失敗
	ErrEmptyNameId                ErrorCode = 0x80000029 // NameIdが未指定
	ErrInvalidNameId              ErrorCode = 0x80000030 // アサーションに格納されたNameIdとセッションに格納されたEmailが不一致
	ErrEmptyLogoutRequestId       ErrorCode = 0x80000031 // LogoutRequest.IDが未指定
	ErrOperationNotAllow          ErrorCode = 0x80000032 // 許可されていない操作

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
	ErrCode    ErrorCode
	Arguments  []interface{}
	Stacktrace string
	cause      error
}

func NewFxtError(errorCode ErrorCode, arguments ...interface{}) *FxtError {
	// スタックトレースの取得
	stack := make([]byte, 4*1024)
	length := runtime.Stack(stack, true)
	stack = stack[:length]

	return &FxtError{
		ErrCode:    errorCode,
		Arguments:  arguments,
		Stacktrace: string(stack),
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
	arguments := strings.Join(common.ArrayMap(func(input interface{}) string {
		return fmt.Sprintf("\"%v\"", input)
	}, e.Arguments), ",")
	return fmt.Sprintf("{\"code\": \"0x%x\", \"arguments\": [%s], \"case\": \"%s\"}", e.ErrCode, arguments, e.cause)
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
				statusCode, errObject := MakeErrorResponse(c, err)
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

func MakeErrorResponse(ctx echo.Context, err error) (int, *gen.Error) {
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
		ctx.Echo().Logger.Errorf("caught invalid Code: %v", fxtError)

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
