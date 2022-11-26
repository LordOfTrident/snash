package token

import (
	"fmt"

	"github.com/LordOfTrident/snash/internal/utils"
)

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
	Word
	BareWord
	Separator

	Help
	Exit
	Echo
	Cd

	Let
	Export

	And
	Or
	Equals

	Error
	count // Count of all token types
)

func (type_ Type) String() string {
	if count != 15 {
		panic("Cover all token types")
	}

	switch type_ {
	case EOF: return "end of file"

	case Integer:   return "integer"
	case Word:      return "string"
	case BareWord:  return "string"
	case Separator: return "separator"

	case Help: return "keyword help"
	case Exit: return "keyword exit"
	case Echo: return "keyword echo"
	case Cd:   return "keyword cd"

	case Let:    return "keyword let"
	case Export: return "keyword export"

	case And:    return "&&"
	case Or:     return "||"
	case Equals: return "="

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

func (tok Token) String() string {
	switch tok.Type {
	case Separator: return "separator (';' or new line)"
	case EOF:       return "end of file"
	case Equals, And, Or: return utils.Quote(tok.Type.String())

	default: return fmt.Sprintf("%v of type %v",
	                            utils.Quote(tok.Data), utils.Quote(tok.Type.String()))
	}
}

func (tok Token) IsKeyword() bool {
	switch tok.Type {
	case Exit, Echo, Cd, Help,
	     Let,  Export: return true

	default: return false
	}
}

func (tok Token) IsStatementEnd() bool {
	return tok.Type == Separator || tok.Type == EOF
}

func (tok Token) IsBinOp() bool {
	return tok.Type == And || tok.Type == Or
}

func (tok Token) IsArgsEnd() bool {
	return tok.IsBinOp() || tok.IsStatementEnd()
}

func (tok Token) IsOp() bool {
	switch tok.Type {
	case Equals: return true

	default: return tok.IsBinOp()
	}
}

func (tok Token) IsString() bool {
	return tok.Type == Word || tok.Type == BareWord
}

func (tok Token) IsArg() bool {
	switch tok.Type {
	case Word, BareWord, Integer: return true

	default: return false
	}
}
