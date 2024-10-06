// Package common 共通機能をまとめたパッケージ
package common

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Code-Hex/synchro/iso8601"
)

var ErrUnknownTimeFormat = errors.New("unknown time format")

func ToTime(v string) (*time.Time, error) {
	t, err := iso8601.ParseDateTime(v)
	if err != nil {
		matches := RegexMT4Date.FindStringSubmatch(v)
		if len(matches) <= 0 {
			fmt.Println("unknown format: ", v)
			return nil, ErrUnknownTimeFormat
		}

		year, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		month, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}
		day, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, err
		}
		hour, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil, err
		}
		min, err := strconv.Atoi(matches[5])
		if err != nil {
			return nil, err
		}
		t = time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)
	}
	return &t, nil
}
