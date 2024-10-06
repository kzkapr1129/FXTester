package lang

import (
	"errors"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/gen"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

// ErrorCode エラーコード
type ErrorCode uint32

/**
  エラーコードの構成について:

  0x8abbcccc の形式でエラーコードを構成します。

  - a: エラーの原因 (1桁目)
    - 1: サーバー起因のエラー
    - 0: クライアント起因のエラー

  - b: エラーメッセージのタイプ (2桁目)
    - 00: インターナルエラー
	- 01: バッドリクエストエラー
    - その他の値は、随時追加されるタイプ

  - c: エラーの詳細番号 (3桁目以降)
*/

const (
	// サーバー起因のエラー
	ErrCodePanic                  ErrorCode = 0x80000001
	ErrCodeUnknownErrorObject     ErrorCode = 0x80000002 // 不明なエラーオブジェクト
	ErrCodeConfig                 ErrorCode = 0x80000003 // 設定ファイル不備
	ErrCodeDisk                   ErrorCode = 0x80000004 // ファイル読み込みエラー
	ErrInvalidIdpMetadata         ErrorCode = 0x80000005 // idPのメタデータの解析に失敗した場合
	ErrDownloadIdpMetadata        ErrorCode = 0x80000006 // idpのメタデータのダウンロードに失敗した場合
	ErrSSOAuthnRequest            ErrorCode = 0x80000007 // SSOのAuthnRequest作成に失敗した場合
	ErrSSOHtmlWriting             ErrorCode = 0x80000008 // SSOのHTML書き込み中のエラー
	ErrCookieNone                 ErrorCode = 0x80000009 // クッキーが見つからなかった場合のエラー
	ErrCodeSSOParseResponse       ErrorCode = 0x80000010 // SAMLレスポンスのパースに失敗した場合のエラー
	ErrRequestParse               ErrorCode = 0x80000011 // parseFormの呼び出しエラー
	ErrUnexpectedAssertion        ErrorCode = 0x80000012 // 予期しないSAMLアサーションを取得した場合のエラー
	ErrDBOpen                     ErrorCode = 0x80000013 // DBのOpenエラー
	ErrDBBegin                    ErrorCode = 0x80000014 // DBのトランザクション開始エラー
	ErrDBRollback                 ErrorCode = 0x80000015 // DBのロールバック失敗
	ErrDBCommit                   ErrorCode = 0x80000016 // DBのコミット失敗
	ErrDBQuery                    ErrorCode = 0x80000017 // DBのクエリエラー
	ErrDBQueryResult              ErrorCode = 0x80000018 // DBのクエリ結果のエラー
	ErrSession                    ErrorCode = 0x80000019 // 不正なセッションエラー (TODO クライアントエラー化)
	ErrSLOAuthnRequest            ErrorCode = 0x80000020 // SLOのAuthnRequest作成に失敗した場合
	ErrSLOValidation              ErrorCode = 0x80000021 // SLOのSAMLResponseのバリデーションに失敗した場合
	ErrJWTSign                    ErrorCode = 0x80000022 // JWTのSignに失敗した場合
	ErrBase64SamlRequest          ErrorCode = 0x80000023 // SAMLリクエストのbase64デコード失敗
	ErrBase64SamlResponse         ErrorCode = 0x80000024 // SAMLレスポンスのbase64デコード失敗
	ErrUnmarshalSamlRequest       ErrorCode = 0x80000025 // SAMLリクエストのUnmarshal失敗
	ErrUnmarshalSamlResponse      ErrorCode = 0x80000026 // SAMLリクエストのUnmarshal失敗
	ErrSamlLogoutResponseCreation ErrorCode = 0x80000027 // SAMLログアウトレスポンスの作成失敗
	ErrEmptyNameId                ErrorCode = 0x80000028 // NameIdが未指定
	ErrInvalidNameId              ErrorCode = 0x80000029 // アサーションに格納されたNameIdとセッションに格納されたEmailが不一致
	ErrEmptyLogoutRequestId       ErrorCode = 0x80000030 // LogoutRequest.IDが未指定
	ErrOperationNotAllow          ErrorCode = 0x80000031 // 許可されていない操作

	// ユーザ起因のエラー
	ErrCodeForbiddenCharacterError ErrorCode = 0x81010001 // 禁止文字エラー
	ErrCodeParameterMissing        ErrorCode = 0x81010002 // 必須パラメータの未指定
	ErrInvalidParameterError       ErrorCode = 0x81010003 // パラメータに予期しない値が設定された場合のエラー
	ErrTooLargeMessageError        ErrorCode = 0x81010004 // multipart/formで巨大なサイズのデータがアップロードされた場合のエラー
	ErrInvalidRequestProtocol      ErrorCode = 0x81010005 // リクエスト形式に不備があった場合のエラー
)

type ErrorTypeDetail struct {
	errorCodePattern *regexp.Regexp
	statusCode       int
	dictKey          string
	displayErrorCode bool
}

var errorTypeDetails = []ErrorTypeDetail{
	{
		errorCodePattern: regexp.MustCompile("0x80[a-zA-Z0-9]{6}"),
		statusCode:       http.StatusInternalServerError,
		dictKey:          "InternalServerError",
		displayErrorCode: true,
	},
	{
		errorCodePattern: regexp.MustCompile(fmt.Sprintf("0x%x", ErrCodeForbiddenCharacterError)),
		statusCode:       http.StatusBadRequest,
		dictKey:          "ForbiddenCharacterError",
		displayErrorCode: true,
	},
	{
		errorCodePattern: regexp.MustCompile(fmt.Sprintf("0x%x", ErrCodeParameterMissing)),
		statusCode:       http.StatusBadRequest,
		dictKey:          "MissingParameterError",
		displayErrorCode: true,
	},
	{
		errorCodePattern: regexp.MustCompile(fmt.Sprintf("0x%x", ErrInvalidParameterError)),
		statusCode:       http.StatusBadRequest,
		dictKey:          "InvalidParameterError",
		displayErrorCode: true,
	},
	{
		errorCodePattern: regexp.MustCompile(fmt.Sprintf("0x%x", ErrTooLargeMessageError)),
		statusCode:       http.StatusBadRequest,
		dictKey:          "TooLargeMessageError",
		displayErrorCode: true,
	},
	{
		errorCodePattern: regexp.MustCompile(fmt.Sprintf("0x%x", ErrInvalidRequestProtocol)),
		statusCode:       http.StatusBadRequest,
		dictKey:          "InvalidRequestProtocolError",
		displayErrorCode: true,
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
	return fmt.Sprintf("{\"code\": \"0x%x\", \"arguments\": [%s], \"case\": \"%v\"}", e.ErrCode, arguments, e.cause)
}

// FindFxtError 原因となったFxtErrorを返却します
func FindFxtError(err error) *FxtError {
	type causer interface {
		Cause() error
	}

	var lastError *FxtError
	for err != nil {
		// Unwrapで取り出せる原初のFxtErrorを探す
		if errors.As(err, &lastError) {
			err = lastError
		}

		// Causeで取り出せるエラーを探す
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return lastError
}

// ErrorHandler Error型エラーがEchoハンドラーから返却された時、エラーメッセージを多言語化しクライアントに返却する
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {
			defer func() {
				if err := recover(); err != nil {
					// スタックトレースのログ出力
					c.Logger().Error(stackTrace())

					// エラーレスポンス返却
					returnErr = c.JSON(http.StatusInternalServerError, &gen.Error{
						Code: uint32(ErrCodePanic),
						Message: GetDict(c, []string{
							"messages",
							"InternalServerError"}, GetDicts(c, []interface{}{ErrCodePanic})...),
					})
				}
			}()

			err := next(c)
			if err != nil {
				// エラーレスポンス返却
				statusCode, errObject := ConvertToGenError(c, err)
				return c.JSON(statusCode, errObject)
			}
			return nil
		}
	}
}

func stackTrace() string {
	stack := make([]byte, 4*1024)
	length := runtime.Stack(stack, true)
	return string(stack[:length])
}

func ConvertToGenError(ctx echo.Context, err error) (int, *gen.Error) {
	fxtError := FindFxtError(err)
	if fxtError == nil {
		// 不明なエラーオブジェクトの場合
		return http.StatusInternalServerError, &gen.Error{
			Code:    uint32(ErrCodeUnknownErrorObject),
			Message: err.Error(),
		}
	}

	errCode := fxtError.ErrCode
	errCodeString := fmt.Sprintf("0x%x", errCode)
	arguments := fxtError.Arguments

	// エラーコードから対応するエラータイプを探す
	for _, errTypeDetail := range errorTypeDetails {
		if errTypeDetail.errorCodePattern.MatchString(errCodeString) {
			// 対応するエラータイプが見つかった場合
			if arguments == nil {
				arguments = []interface{}{}
			}
			if errTypeDetail.displayErrorCode {
				arguments = append(arguments, errCode)
			}
			return errTypeDetail.statusCode, &gen.Error{
				Code: uint32(errCode),
				Message: GetDict(ctx, []string{
					"messages",
					errTypeDetail.dictKey}, GetDicts(ctx, arguments)...),
			}
		}
	}

	// エラーコードが全てのエラータイプに一致しなかった場合
	return http.StatusInternalServerError, &gen.Error{
		Code: uint32(errCode),
		Message: GetDict(ctx, []string{
			"messages",
			"InternalServerError"}, GetDicts(ctx, []interface{}{errCode})...),
	}
}
