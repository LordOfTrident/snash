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

// Help

type HelpStatement struct {
	Token token.Token

	Path string
}

func (hs *HelpStatement) statementNode() {}

func (hs *HelpStatement) NodeToken() token.Token {
	return hs.Token
}

func (hs *HelpStatement) NodeTypeToString() string {
	return "help statement"
}

// Binary operator

type BinOpStatement struct {
	Token token.Token

	Left, Right Statement
}

func (bin *BinOpStatement) statementNode() {}

func (bin *BinOpStatement) NodeToken() token.Token {
	return bin.Token
}

func (bin *BinOpStatement) NodeTypeToString() string {
	return "binary operator " + bin.NodeToken().Type.String()
}

// Variables

type LetStatement struct {
	Token token.Token

	Name, Value string
}

func (let *LetStatement) statementNode() {}

func (let *LetStatement) NodeToken() token.Token {
	return let.Token
}

func (let *LetStatement) NodeTypeToString() string {
	return "let statement"
}

type AssignStatement struct {
	Token token.Token

	Name, Value string
}

func (as *AssignStatement) statementNode() {}

func (as *AssignStatement) NodeToken() token.Token {
	return as.Token
}

func (as *AssignStatement) NodeTypeToString() string {
	return "assign statement"
}

type ExportStatement struct {
	Token token.Token

	Name string
}

func (export *ExportStatement) statementNode() {}

func (export *ExportStatement) NodeToken() token.Token {
	return export.Token
}

func (export *ExportStatement) NodeTypeToString() string {
	return "export statement"
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
