package token

import "fmt"

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

	typeCount
)

func (type_ Type) String() string {
	if typeCount != 5 {
		panic("Cover all token types")
	}

	switch type_ {
	case EOF: return "end of file"

	case String:    return "string"
	case Integer:   return "integer"
	case Separator: return "separator"

	case KeywordExit: return "keyword exit"

	default: panic("Unreachable")
	}
}

type Token struct {
	Type Type
	Data string

	Where Where
}

func New(type_ Type, data string, where Where) Token {
	return Token{Type: type_, Data: data, Where: where}
}

func Empty() Token {
	return Token{Type: EOF}
}

func (tok *Token) IsStatementEnd() bool {
	return tok.Type == Separator || tok.Type == EOF
}
