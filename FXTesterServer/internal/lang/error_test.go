package lang

import (
	"encoding/json"
	"errors"
	"fmt"
	"fxtester/internal/gen"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Test_ErrorHandler(t *testing.T) {
	type args struct {
		ctx  func(w http.ResponseWriter) echo.Context
		next echo.HandlerFunc
	}

	newNoLoggerEcho := func() *echo.Echo {
		e := echo.New()
		e.Logger.SetLevel(log.OFF)
		return e
	}

	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantBody         bool
		wantErrorCode    ErrorCode
		wantErrorMessage string
	}{
		{
			name: "test1",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return nil
				},
			},
			wantErr:  false,
			wantBody: false,
		},
		{
			name: "test2",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return errors.New("test-error")
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeUnknownErrorObject,
			wantErrorMessage: "インターナルエラーが発生しました。\n(エラーコード: 0x80000002)",
		},
		{
			name: "test3",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return NewFxtError(100, "test")
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeUnknownErrorCode,
			wantErrorMessage: "インターナルエラーが発生しました。\n(エラーコード: 0x80000003)",
		},
		{
			name: "test4",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return NewFxtError(ErrCodeForbiddenCharacterError, "test")
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeForbiddenCharacterError,
			wantErrorMessage: "testに禁止文字が指定されました。\n(エラーコード: 0x81010001)",
		},
		{
			name: "test5",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return NewFxtError(ErrCodeForbiddenCharacterError, "test").SetCause(NewFxtError(ErrCodeForbiddenCharacterError, "test2"))
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeForbiddenCharacterError,
			wantErrorMessage: "test2に禁止文字が指定されました。\n(エラーコード: 0x81010001)",
		},
		{
			name: "test6",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return NewFxtError(ErrCodeForbiddenCharacterError, "test").SetCause(fmt.Errorf("test-error"))
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeForbiddenCharacterError,
			wantErrorMessage: "testに禁止文字が指定されました。\n(エラーコード: 0x81010001)",
		},
		{
			name: "test7",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					return NewFxtError(ErrCodeForbiddenCharacterError, "words.name").SetCause(fmt.Errorf("test-error"))
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodeForbiddenCharacterError,
			wantErrorMessage: "名前に禁止文字が指定されました。\n(エラーコード: 0x81010001)",
		},
		{
			name: "test8",
			args: args{
				ctx: func(w http.ResponseWriter) echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					return newNoLoggerEcho().NewContext(req, w)
				},
				next: func(c echo.Context) error {
					panic("test-panic")
				},
			},
			wantErr:          false,
			wantBody:         true,
			wantErrorCode:    ErrCodePanic,
			wantErrorMessage: "インターナルエラーが発生しました。\n(エラーコード: 0x80000001)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			if err := ErrorHandler()(tt.args.next)(tt.args.ctx(w)); (err != nil) != tt.wantErr {
				t.Errorf("ErrorHandler()()()=%v want=%v", err, tt.wantErr)
			} else if tt.wantBody != (0 < len(w.Body.Bytes())) {
				t.Errorf("ErrorHandler()()()=%v wantBody=%v", w.Body.String(), tt.wantBody)
			} else if tt.wantBody && (0 < len(w.Body.Bytes())) {
				var res gen.Error
				if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
					t.Errorf("Error in json.Unmarshal: %v", err)
				} else if tt.wantErrorCode != ErrorCode(res.Code) {
					t.Errorf("ErrorHandler()()()=%v wantErrorCode=0x%x", res.Code, tt.wantErrorCode)
				} else if tt.wantErrorMessage != res.Message {
					t.Errorf("ErrorHandler()()()=%v wantErrorMessage=%v", res.Message, tt.wantErrorMessage)
				}
			}

		})
	}
}
