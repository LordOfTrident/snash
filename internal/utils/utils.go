package utils

import (
	"strings"
)

func Unescape(str string) string {
	str = strings.Replace(str, "\\",       "\\\\", -1)
	str = strings.Replace(str, string(27), "\\e",  -1)
	str = strings.Replace(str, "\n",       "\\n",  -1)
	str = strings.Replace(str, "\r",       "\\r",  -1)
	str = strings.Replace(str, "\t",       "\\t",  -1)
	str = strings.Replace(str, "\v",       "\\v",  -1)
	str = strings.Replace(str, "\b",       "\\b",  -1)
	str = strings.Replace(str, "\f",       "\\f",  -1)
	str = strings.Replace(str, "\"",       "\\\"", -1)

	return str
}

func Quote(str string) string {
	str = Unescape(str)

	return "\"" + str + "\""
}
