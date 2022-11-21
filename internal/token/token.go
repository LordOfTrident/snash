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

	Integer
	String
	Separator

	LParen
	RParen

	Exit
	Echo
	Cd

	And
	Or

	Error
	count // Count of all token types
)

func (type_ Type) String() string {
	if count != 12 {
		panic("Cover all token types")
	}

	switch type_ {
	case EOF: return "end of file"

	case Integer:   return "integer"
	case String:    return "string"
	case Separator: return "separator"

	case LParen: return "("
	case RParen: return ")"

	case Exit: return "keyword exit"
	case Echo: return "keyword echo"
	case Cd:   return "keyword cd"

	case And: return "&&"
	case Or:  return "||"

	case Error: return "error"

	default: panic("Unreachable")
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
