package validator

import (
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/gen"
)

func ValidateCandle(candle gen.Candle) error {
	// 日付の文字列フォーマットをチェックする
	if !common.RegexISO8601.MatchString(candle.Time) {
		return fmt.Errorf("invalid time format: %v", candle.Time)
	}

	// 数値の範囲チェック
	if candle.Open < 0.0 || candle.High < 0.0 || candle.Low < 0.0 || candle.Close < 0.0 {
		return fmt.Errorf("invalid number: %f,%f,%f,%f", candle.Open, candle.High, candle.Low, candle.Close)
	}

	// 数値の論理性チェック (高値)
	if candle.High < candle.Low || candle.High < candle.Open || candle.High < candle.Close {
		return fmt.Errorf("invalid high: %f", candle.High)
	}

	// 数値の論理性チェック (安値)
	if candle.Open < candle.Low || candle.Close < candle.Low {
		return fmt.Errorf("invalid low: %f", candle.Low)
	}

	return nil
}
