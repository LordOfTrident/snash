package errors

import (
	"fmt"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/token"
	"github.com/LordOfTrident/snash/internal/node"
)

type Error struct {
	Where token.Where
	Msg   string
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Where, err.Msg)
}

func New(where token.Where, format string, args... interface{}) error {
	return Error{Where: where, Msg: fmt.Sprintf(format, args...)}
}

func UnexpectedToken(tok *token.Token) error {
	return New(tok.Where, "Unexpected %v", tok)
}

func ExpectedToken(tok *token.Token, expected token.Type) error {
	return New(tok.Where, "Expected type %v, got %v",
	           utils.Quote(expected.String()), tok)
}

func CmdNotFound(cmd string, where token.Where) error {
	return New(where, "Command %v not found", utils.Quote(cmd))
}

func FileNotFound(path string, where token.Where) error {
	return New(where, "File/directory %v not found", utils.Quote(path))
}

func VarNotFound(name string, where token.Where) error {
	return New(where, "Variable %v not found", utils.Quote(name))
}

func UnexpectedNode(node node.Node) error {
	return New(node.NodeToken().Where, "Unexpected %v", node.NodeToken())
}

func ErrorTokenToError(tok token.Token) error {
	return fmt.Errorf("%v: %v", tok.Where, tok.Data)
}
