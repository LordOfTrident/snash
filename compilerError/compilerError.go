package compilerError

import (
	"fmt"

	"github.com/LordOfTrident/snash/token"
	"github.com/LordOfTrident/snash/node"
)

type Error struct {
	Where token.Where
	Msg   string
}

func tokenToErrorStr(tok *token.Token) string {
	switch tok.Type {
	case token.Separator: return "separator (';' or new line)"

	default: return fmt.Sprintf("'%v' of type '%v'", tok.Data, tok.Type)
	}
}

func nodeToErrorStr(node node.Node) string {
	tok := node.NodeToken()

	switch tok.Type {
	case token.Separator: return "separator (';' or new line)"

	default: return fmt.Sprintf("'%v' of type '%v'", tok.Data, node.NodeTypeToString())
	}
}

func New(where token.Where, format string, args... interface{}) Error {
	return Error{Where: where, Msg: fmt.Sprintf(format, args...)}
}

func UnexpectedToken(tok *token.Token) Error {
	return New(tok.Where, "Unexpected %v", tokenToErrorStr(tok))
}

func UnexpectedNode(node node.Node) Error {
	return New(node.NodeToken().Where, "Unexpected %s", nodeToErrorStr(node))
}

func ExpectedToken(tok *token.Token, expected token.Type) Error {
	return New(tok.Where, "Expected type '%v', got %s", expected, tokenToErrorStr(tok))
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %s", err.Where, err.Msg)
}
