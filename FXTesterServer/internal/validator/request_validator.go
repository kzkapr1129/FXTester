package validator

import (
	"encoding/json"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"slices"

	"github.com/labstack/echo/v4"
)

func ValidatePostZigzag(ctx echo.Context) error {

	inputDataTypes := ctx.Request().MultipartForm.Value["type"]
	csvInfos := ctx.Request().MultipartForm.Value["csvInfo"]
	csvs := ctx.Request().MultipartForm.File["csv"]
	candless := ctx.Request().MultipartForm.Value["candles"]

	// 入力タイプが'csv'と'candles'の個数をカウントする
	numInputTypeCsv, numInputTypeCandles, err := func() (int, int, error) {
		numInputTypeCsv := 0
		numInputTypeCandles := 0
		for _, inputType := range inputDataTypes {
			switch inputType {
			case string(gen.PostZigzagRequestTypeCsv):
				numInputTypeCsv++
			case string(gen.PostZigzagRequestTypeCandles):
				numInputTypeCandles++
			default:
				return 0, 0, lang.NewFxtError(lang.ErrInvalidParameterError, "type")
			}
		}
		return numInputTypeCsv, numInputTypeCandles, nil
	}()
	if err != nil {
		return err
	}

	// 文字列配列の中で空文字以外の要素の数をカウントする
	countNotEmpty := func(arr []string) int {
		count := 0
		for _, v := range arr {
			if v != "" {
				count++
			}
		}
		return count
	}

	// 'type'パラメータの未指定チェック'
	if numInputTypeCsv == 0 && numInputTypeCandles == 0 {
		return lang.NewFxtError(lang.ErrCodeParameterMissing, "type")
	}

	// 'csvInfo'パラメータの個数チェック
	if numInputTypeCsv != countNotEmpty(csvInfos) {
		return lang.NewFxtError(lang.ErrInvalidParameterError, "csvInfo")
	}

	// 'csv'パラメータの個数チェック
	if numInputTypeCsv != len(csvs) {
		return lang.NewFxtError(lang.ErrInvalidParameterError, "csv")
	}

	// 'candles'パラメータの個数チェック
	if numInputTypeCandles != countNotEmpty(candless) {
		return lang.NewFxtError(lang.ErrInvalidParameterError, "candles")
	}

	for i, v := range csvInfos {
		if v == "" {
			// multipartの動作上、空文字が指定されることがある
			continue
		}

		var t gen.CsvInfo

		// unmarshalが可能かチェックする
		if err := json.Unmarshal([]byte(v), &t); err != nil {
			return lang.NewFxtError(lang.ErrInvalidParameterError, fmt.Sprintf("csvInfo[%d]", i)).SetCause(err)
		}

		indexes := []int{t.CloseColumnIndex, t.HighColumnIndex, t.LowColumnIndex, t.OpenColumnIndex, t.TimeColumnIndex}
		slices.Sort(indexes)
		unique := slices.Compact(indexes)

		// インデックスの重複チェック
		if len(unique) != len(indexes) {
			return lang.NewFxtError(lang.ErrInvalidParameterError, fmt.Sprintf("csvInfo[%d]", i))
		}

		// 区切り文字のチェック
		if t.DelimiterChar == "" || !common.RegexCsvDelimiter.MatchString(string(t.DelimiterChar)) {
			return lang.NewFxtError(lang.ErrInvalidParameterError, fmt.Sprintf("csvInfo[%d]", i))
		}
	}

	for i, v := range candless {
		if v == "" {
			// multipartの動作上、空文字が指定されることがある
			continue
		}

		var ts []gen.Candle

		// unmarshalが可能かチェックする
		if err := json.Unmarshal([]byte(v), &ts); err != nil {
			return lang.NewFxtError(lang.ErrInvalidParameterError, fmt.Sprintf("candles[%d]", i)).SetCause(err)
		}

		for j, t := range ts {
			// Candle型のバリデーション
			if err := ValidateCandle(t); err != nil {
				return lang.NewFxtError(lang.ErrInvalidParameterError, fmt.Sprintf("candles[%d][%d]", i, j)).SetCause(err)
			}
		}
	}

	return nil
}
