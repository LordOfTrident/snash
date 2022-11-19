package token

import "fmt"

// Data about where the token was

type Where struct {
	Row, Col int
	Path     string
}

func (where Where) String() string {
	return fmt.Sprintf("%v:%v:%v", where.Path, where.Row, where.Col)
}

type Type int
const (
	EOF = iota

	String
	Integer
	Separator

	KeywordExit
	KeywordEcho
	KeywordCd

	Error
	count // Count of all token types
)

func (type_ Type) String() string {
	if count != 8 {
		panic("Cover all token types")
	}

	switch type_ {
	case EOF: return "end of file"

	case String:    return "string"
	case Integer:   return "integer"
	case Separator: return "separator"

	case KeywordExit: return "keyword exit"
	case KeywordEcho: return "keyword echo"
	case KeywordCd:   return "keyword cd"

	case Error: return "error"

	default: panic("Unreachable")
	}
}

func (type_ Type) IsKeyword() bool {
	switch type_ {
	case KeywordExit, KeywordEcho, KeywordCd: return true

	default: return false
	}
}

type Token struct {
	Type   Type
	Data   string
	TxtLen int

	Where Where
}

func New(type_ Type, data string, where Where, txtLen int) Token {
	return Token{Type: type_, Data: data, Where: where, TxtLen: txtLen}
}

func NewError(where Where, txtLen int, format string, args... interface{}) Token {
	return New(Error, fmt.Sprintf(format, args...), where, txtLen)
}

func NewEOF(where Where) Token {
	return New(EOF, "", where, 0)
}
