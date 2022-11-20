package node

import "github.com/LordOfTrident/snash/internal/token"

// Base node interface

type Node interface {
	NodeToken()        token.Token
	NodeTypeToString() string
}

// Statements

type Statement interface {
	Node

	statementNode()
}

// Exit

type ExitStatement struct {
	Token token.Token

	Ex    int
	HasEx bool
}

func (es *ExitStatement) statementNode() {}

func (es *ExitStatement) NodeToken() token.Token {
	return es.Token
}

func (es *ExitStatement) NodeTypeToString() string {
	return "exit statement"
}

// Echo

type EchoStatement struct {
	Token token.Token

	Msg string
}

func (echo *EchoStatement) statementNode() {}

func (echo *EchoStatement) NodeToken() token.Token {
	return echo.Token
}

func (echo *EchoStatement) NodeTypeToString() string {
	return "echo statement"
}

// Cd

type CdStatement struct {
	Token token.Token

	Path string
}

func (cd *CdStatement) statementNode() {}

func (cd *CdStatement) NodeToken() token.Token {
	return cd.Token
}

func (cd *CdStatement) NodeTypeToString() string {
	return "cd statement"
}

// And

const (
	LogicalAnd = iota
	LogicalOr
)

type LogicalOpStatement struct {
	Token token.Token

	Type        int
	Left, Right Statement
}

func (lo *LogicalOpStatement) statementNode() {}

func (lo *LogicalOpStatement) NodeToken() token.Token {
	return lo.Token
}

func (lo *LogicalOpStatement) NodeTypeToString() string {
	return "logical operator statement"
}

// Command

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

// Statements

type Statements struct {
	List []Statement
}

func (s *Statements) NodeToken() token.Token {
	if len(s.List) == 0 {
		return token.NewEOF(token.Where{})
	}

	return s.List[0].NodeToken()
}

func (s *Statements) NodeTypeToString() string {
	return "statements"
}
