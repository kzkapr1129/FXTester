package common

import "regexp"

// RegexCsvDelimiter .csvの区切り文字
var RegexCsvDelimiter = regexp.MustCompile(`^[,\t\s]$`)

// RegexISO8601 ISO8601フォーマットの日付
var RegexISO8601 = regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[1-2][0-9]|3[0-1])T(?:[0-1][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](?:\.[0-9]+)?(?:Z|[+-](?:[0-1][0-9]|2[0-3]):[0-5][0-9])$`)
