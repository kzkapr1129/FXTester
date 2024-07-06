package internal

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func ArrayMapSafe[INPUT any, OUTPUT any](conv func(v INPUT) (OUTPUT, error), values []INPUT) ([]OUTPUT, error) {
	out := []OUTPUT{}
	for _, v := range values {
		o, err := conv(v)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func ArrayMap[INPUT any, OUTPUT any](conv func(v INPUT) OUTPUT, values []INPUT) []OUTPUT {
	out := []OUTPUT{}
	for _, v := range values {
		out = append(out, conv(v))
	}
	return out
}

func GetEnvAs[T uint16 | int | string](envName string, required bool, defaultValue T) (T, error) {
	var v T
	var envValue string
	if envValue = os.Getenv(envName); required && envValue == "" {
		return v, fmt.Errorf("make sure '%s' is specified in the environment variable", envName)
	} else if envValue == "" {
		return defaultValue, nil
	}

	switch (interface{})(v).(type) {
	case uint16:
		if intValue, err := strconv.Atoi(envValue); err != nil {
			return v, err
		} else {
			return (interface{})(uint16(intValue)).(T), nil
		}
	case int:
		if intValue, err := strconv.Atoi(envValue); err != nil {
			return v, err
		} else {
			return (interface{})(intValue).(T), nil
		}
	case string:
		return (interface{})(envValue).(T), nil
	default:
		return v, errors.New("unknown type")
	}
}
