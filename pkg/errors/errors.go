package errors

import (
	"fmt"

	"github.com/LordOfTrident/snash/pkg/utils"
	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/node"
)

type Error struct {
	Where token.Where
	Msg   string
}

func tokenToString(tok token.Token) string {
	// For outputting better errors

	switch tok.Type {
	case token.Separator: return "separator (';' or new line)"
	case token.EOF:       return "end of file"

	default: return fmt.Sprintf("%v of type %v",
	                            utils.Quote(tok.Data), utils.Quote(tok.Type.String()))
	}
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Where, err.Msg)
}

func New(where token.Where, format string, args... interface{}) error {
	return Error{Where: where, Msg: fmt.Sprintf(format, args...)}
}

func UnexpectedToken(tok *token.Token) error {
	return New(tok.Where, "Unexpected %v", tokenToString(*tok))
}

func ExpectedToken(tok *token.Token, expected token.Type) error {
	return New(tok.Where, "Expected type %v, got %v",
	           utils.Quote(expected.String()), tokenToString(*tok))
}

func CmdNotFound(cmd string, where token.Where) error {
	return New(where, "Command %v not found", utils.Quote(cmd))
}

func FileNotFound(path string, where token.Where) error {
	return New(where, "File/directory %v not found", utils.Quote(path))
}

func UnexpectedNode(node node.Node) error {
	return New(node.NodeToken().Where, "Unexpected %v", tokenToString(node.NodeToken()))
}

func ErrorTokenToError(tok token.Token) error {
	return fmt.Errorf("%v: %v", tok.Where, tok.Data)
}
