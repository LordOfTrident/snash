package errors

import (
	"fmt"

	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/node"
)

type Error struct {
	Where token.Where
	Msg   string
}

func unescape(str string) (ret string) {
	for _, ch := range str {
		switch ch {
		case 27:   ret += "'$'\\e"
		case '\n': ret += "'$'\\n"
		case '\r': ret += "'$'\\r"
		case '\t': ret += "'$'\\t"
		case '\v': ret += "'$'\\v"
		case '\b': ret += "'$'\\b"
		case '\f': ret += "'$'\\f"

		default: ret += string(ch)
		}
	}

	return
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Where, err.Msg)
}

func New(where token.Where, format string, args... interface{}) error {
	return Error{Where: where, Msg: unescape(fmt.Sprintf(format, args...))}
}

func UnexpectedToken(tok *token.Token) error {
	return New(tok.Where, "Unexpected %v", tok)
}

func ExpectedToken(tok *token.Token, expected token.Type) error {
	return New(tok.Where, "Expected type '%v', got %v", expected, tok)
}

func CmdNotFound(cmd string, where token.Where) error {
	return New(where, "Command '%v' not found", cmd)
}

func FileNotFound(path string, where token.Where) error {
	return New(where, "File/directory '%v' not found", path)
}

func UnexpectedNode(node node.Node) error {
	return New(node.NodeToken().Where, "Unexpected %v", node.NodeToken())
}

func ErrorTokenToError(tok token.Token) error {
	return fmt.Errorf("%v: %v", tok.Where, tok.Data)
}
