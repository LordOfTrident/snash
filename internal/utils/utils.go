package utils

import (
	"strings"
)

const CharNone byte = 0

func IsWhitespace(char byte) bool {
	switch char {
	case ' ', '\r', '\t', '\v', '\f': return true

	default: return false
	}
}

func IsSeparator(char byte) bool {
	switch char {
	case ' ', '\r', '\t', '\n', '\v', '\f', ';': return true

	default: return false
	}
}

func IsAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func IsDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func IsAlphanum(char byte) bool {
	return IsAlpha(char) || IsDigit(char)
}

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
