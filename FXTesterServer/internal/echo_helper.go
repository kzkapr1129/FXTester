package internal

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

var echo_helper = struct {
	re *regexp.Regexp
}{
	re: regexp.MustCompile(`((?:[a-zA-Z]+|\*)(?:-[a-zA-Z]+)?)(?:\s*;\s*q\s*=\s*([0-9]\.[0-9]))?`),
}

func GetLocales(ctx echo.Context) []string {
	defaultLocales := []string{"ja"}

	if v := ctx.Request().Header.Get("Accept-Language"); v != "" {
		acceptLanguages := ArrayMapSkip(func(input string) (*struct {
			lang string
			q    float64
		}, bool) {
			if langQ := strings.TrimSpace(input); langQ == "" {
				// Accept-Languageに空文字の設定が格納されている場合
				return nil, true
			} else if langQValues := echo_helper.re.FindStringSubmatch(langQ); len(langQValues) != 3 {
				// 予期しない設定がAccept-Languageに設定されている場合
				return nil, true
			} else if langQValues[2] == "" {
				// 重み値が指定されている場合
				return &struct {
					lang string
					q    float64
				}{
					lang: langQValues[1],
					q:    1,
				}, false
			} else if q, err := strconv.ParseFloat(langQValues[2], 32); err != nil {
				// 重み値に不正な値が設定されている場合
				return nil, true
			} else {
				// 重み値が指定されている場合
				return &struct {
					lang string
					q    float64
				}{
					lang: langQValues[1],
					q:    q,
				}, false
			}

		}, strings.Split(v, ","))

		if len(acceptLanguages) <= 0 {
			return defaultLocales
		}

		sort.Slice(acceptLanguages, func(i, j int) bool {
			iv := acceptLanguages[i]
			jv := acceptLanguages[j]
			if iv.q == jv.q {
				return false
			}
			return iv.q > jv.q
		})

		return ArrayMap(func(input *struct {
			lang string
			q    float64
		}) string {
			return input.lang
		}, acceptLanguages)
	}

	// デフォルトのロケーションを返す
	return defaultLocales
}
