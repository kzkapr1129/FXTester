package common

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

func ArrayMapSkip[INPUT any, OUTPUT any](conv func(v INPUT) (cv OUTPUT, isSkip bool), values []INPUT) []OUTPUT {
	out := []OUTPUT{}
	for _, v := range values {
		if cv, isSkip := conv(v); !isSkip {
			out = append(out, cv)
		}
	}
	return out
}

func Set[T comparable](values []T) []T {
	m := map[T]struct{}{}
	results := []T{}
	for _, v := range values {
		if _, ok := m[v]; !ok {
			results = append(results, v)
			m[v] = struct{}{}
		}
	}
	return results
}

func GetEnvAs[T uint16 | int | string](envName string, required bool, defaultValue T) (T, error) {
	var v T
	var envValue string
	if envValue = os.Getenv(envName); required && envValue == "" {
		return v, fmt.Errorf("make sure '%s' is specified in the environment variable", envName)
	} else if envValue == "" && !required {
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
