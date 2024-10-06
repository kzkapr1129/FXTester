package reader

import (
	"encoding/csv"
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/gen"
	"io"
	"strconv"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func ReadCandleCsv(csvInfo gen.CsvInfo, r io.Reader) (res []common.Candle, lastError error) {

	utf16bom := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	utf8Reader := transform.NewReader(r, utf16bom.NewDecoder())

	reader := csv.NewReader(utf8Reader)
	reader.Comma = []rune(csvInfo.DelimiterChar)[0]
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("failed readAll: ", err)
		return nil, err
	}

	getColValue := func(row, col int) (string, error) {
		if len(records) <= row {
			return "", fmt.Errorf("invalid row: %d", row)
		}
		if len(records[row]) <= col {
			return "", fmt.Errorf("invalid col: %d", col)
		}
		return records[row][col], nil
	}

	candles := []common.Candle{}
	for row := 0; row < len(records); row++ {
		colTime, err := getColValue(row, csvInfo.TimeColumnIndex)
		if err != nil {
			return nil, err
		}
		fmt.Println("colTime: ", colTime)

		colHigh, err := getColValue(row, csvInfo.HighColumnIndex)
		if err != nil {
			return nil, err
		}

		colOpen, err := getColValue(row, csvInfo.OpenColumnIndex)
		if err != nil {
			return nil, err
		}

		colClose, err := getColValue(row, csvInfo.CloseColumnIndex)
		if err != nil {
			return nil, err
		}

		colLow, err := getColValue(row, csvInfo.LowColumnIndex)
		if err != nil {
			return nil, err
		}

		time, err := common.ToTime(colTime)
		fmt.Println(colTime, ": ", time)
		if err != nil {
			return nil, err
		}

		high, err := strconv.ParseFloat(colHigh, 32)
		if err != nil {
			return nil, err
		}
		open, err := strconv.ParseFloat(colOpen, 32)
		if err != nil {
			return nil, err
		}
		close, err := strconv.ParseFloat(colClose, 32)
		if err != nil {
			return nil, err
		}
		low, err := strconv.ParseFloat(colLow, 32)
		if err != nil {
			return nil, err
		}

		candles = append(candles, common.Candle{
			Time:  *time,
			High:  high,
			Open:  open,
			Close: close,
			Low:   low,
		})
	}

	return candles, nil
}
