package lang

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_GetDict(t *testing.T) {
	type args struct {
		ctx       echo.Context
		errCode   string
		parameter []interface{}
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "インターナルエラー(ja)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました。\n(エラーコード: 0x123)",
		},
		{
			name: "インターナルエラー(ja-JP)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja-JP")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました。\n(エラーコード: 0x123)",
		},
		{
			name: "インターナルエラー(*)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "*")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました。\n(エラーコード: 0x123)",
		},
		{
			name: "インターナルエラー(ja:0.5, en)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja;q=0.5, en")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました(要英訳)。\n(エラーコード: 0x123)",
		},
		{
			name: "インターナルエラー(ja:0.5, en:0.3)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					req.Header.Set("Accept-Language", "ja;q=0.5, en;q=0.3")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました。\n(エラーコード: 0x123)",
		},
		{
			name: "インターナルエラー(未指定)",
			args: args{
				ctx: func() echo.Context {
					req := httptest.NewRequest("GET", "http://localhost", nil)
					//req.Header.Set("Accept-Language", "ja;q=0.5, en;q=0.3")
					rec := httptest.NewRecorder()
					return echo.New().NewContext(req, rec)
				}(),
				errCode: "InternalError",
				parameter: []interface{}{
					0x123,
				},
			},
			want: "インターナルエラーが発生しました。\n(エラーコード: 0x123)",
		},
	}

	replaceLR := func(str string) string {
		return strings.ReplaceAll(str, "\n", "\\n")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if v := GetDict(tt.args.ctx, []string{"messages", tt.args.errCode}, tt.args.parameter...); v != tt.want {
				t.Errorf("GetDict()=%v want=%v", replaceLR(v), replaceLR(tt.want))
			}
		})
	}
}
