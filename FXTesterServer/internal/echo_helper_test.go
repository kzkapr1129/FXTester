package internal

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_GetLocales(t *testing.T) {
	type args struct {
		ctx echo.Context
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "en"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"en",
			},
		},
		{
			name: "test2",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "en, de"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"en", "de",
			},
		},
		{
			name: "test3",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "en-US, de, jp;q=0.5"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"en-US", "de", "jp",
			},
		},
		{
			name: "test4",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "en, aa;q=0.5, bb;q=0.6, cc;q=0.4"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"en", "bb", "aa", "cc",
			},
		},
		{
			name: "test5",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "jp, en-US ; q = 1.5"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"en-US", "jp",
			},
		},
		{
			name: "test6",
			args: args{
				ctx: func() echo.Context {
					const acceptLanguage = "1"
					req := httptest.NewRequest("GET", "http://test.co.jp", nil)
					req.Header.Set("Accept-Language", acceptLanguage)
					return echo.New().NewContext(req, nil)
				}(),
			},
			want: []string{
				"ja",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if v := GetLocales(tt.args.ctx); !reflect.DeepEqual(v, tt.want) {
				t.Errorf("GetLocales()=%v want=%v", v, tt.want)
			}
		})
	}
}
