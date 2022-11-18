package errors

import (
	"fmt"
	"os"

	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/node"
	"github.com/LordOfTrident/snash/pkg/attr"
)

func Print(err error) {
	fmt.Fprintf(os.Stderr, "%v%vError:%v %v\n", attr.Bold, attr.BrightRed, attr.Reset, err.Error())
}

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
