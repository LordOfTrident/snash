package node

import "github.com/LordOfTrident/snash/token"

type Node interface {
	NodeToken()        token.Token
	NodeTypeToString() string
}

type Statement interface {
	Node

	statementNode()
}

type ExitStatement struct {
	Token token.Token

	Exitcode int
}

func (es *ExitStatement) statementNode() {}

func (es *ExitStatement) NodeToken() token.Token {
	return es.Token
}

func (es *ExitStatement) NodeTypeToString() string {
	return "exit statement"
}

type CmdStatement struct {
	Token token.Token

	Cmd  string
	Args []string
}

func (cs *CmdStatement) statementNode() {}

func (cs *CmdStatement) NodeToken() token.Token {
	return cs.Token
}

func (cs *CmdStatement) NodeTypeToString() string {
	return "command"
}

type Statements struct {
	List []Statement
}

func (s *Statements) NodeToken() token.Token {
	if len(s.List) == 0 {
		return token.Empty()
	}

	return s.List[0].NodeToken()
}

func (s *Statements) NodeTypeToString() string {
	return "statements"
}
