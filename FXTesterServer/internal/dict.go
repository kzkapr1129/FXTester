package internal

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type dictData struct {
	Data  map[string]interface{} `yaml:"dict"`
	Alias map[string]string      `yaml:"alias"`

	regexes []struct {
		regex *regexp.Regexp
		alias string
	}
}

var dict = struct {
	once sync.Once
	data dictData
}{}

func LoadDict() {
	dict.once.Do(func() {
		fn := GetConfig().DictFilePath
		file, err := os.Open(fn)
		if err != nil {
			panic("failed to open " + fn)
		}
		defer file.Close()

		bytes, err := io.ReadAll(file)
		if err != nil {
			panic("failed to load " + fn)
		}

		if err := yaml.Unmarshal(bytes, &dict.data); err != nil {
			panic("failed to unmarshal " + fn)
		}

		for regex, alias := range dict.data.Alias {
			re := regexp.MustCompile(regex)
			dict.data.regexes = append(dict.data.regexes, struct {
				regex *regexp.Regexp
				alias string
			}{
				regex: re,
				alias: alias,
			})
		}
	})
}

func GetDict(ctx echo.Context, keys []string, arguments ...interface{}) string {
	// 辞書の読み込み(初回のみ)
	LoadDict()

	// 辞書の読み込み失敗時のエラー文字列を生成
	errDict := func() string {
		return fmt.Sprintf("[%s]", strings.Join(keys, "."))
	}()

	// 指定されたキーから辞書のツリーを辿る
	data := dict.data.Data
	for _, key := range keys {
		if tmp, ok := data[key]; !ok {
			// 指定したキーの辞書が存在しない場合
			return errDict
		} else if v, ok := tmp.(map[string]interface{}); ok {
			data = v
		}
	}

	// Accept-Languageからロケーションコードを取得する
	locs := Set(ArrayMap(func(v string) string {
		tmp := v
		for _, r := range dict.data.regexes {
			// Aliasの変換条件に一致する場合はAliasに変換する
			if v = r.regex.ReplaceAllString(v, r.alias); v != tmp {
				// 変換に成功した場合
				return v
			}
		}
		return v
	}, GetLocales(ctx)))

	// ロケーションの優先度順に辞書を調べる
	for _, loc := range locs {
		if d, ok := data[loc]; !ok {
			// ロケーションが存在しない場合
			continue
		} else if v, ok := d.(string); !ok {
			// 辞書に文字列以外が格納されている場合
			return errDict
		} else if 0 < len(arguments) {
			// 辞書発見。引数が指定されている場合
			return trimLB(fmt.Sprintf(v, arguments...))
		} else {
			// 辞書発見
			return trimLB(v)
		}
	}

	// ロケーションが全て一致しない場合
	return errDict
}

func trimLB(input string) string {
	return strings.TrimRight(strings.TrimRight(input, "\r\n"), "\n")
}
