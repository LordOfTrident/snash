package parser

import (
	"strconv"

	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/token"
	"github.com/LordOfTrident/snash/internal/lexer"
	"github.com/LordOfTrident/snash/internal/node"
)

type Parser struct {
	idx int
	tok *token.Token

	Toks []token.Token
}

func isStatementEnd(tok *token.Token) bool {
	return tok.Type == token.Separator || tok.Type == token.EOF
}

func isBinOp(tok *token.Token) bool {
	return tok.Type == token.And || tok.Type == token.Or
}

func isArgsEnd(tok *token.Token) bool {
	switch tok.Type {
	case token.LParen, token.RParen: return true

	default: return isBinOp(tok) || isStatementEnd(tok)
	}
}

func New(source, path string) (*Parser, error) {
	// The parser will automatically lex the source
	l := lexer.New(source, path)

	toks, err := l.Lex()
	if err != nil {
		return nil, err
	}

	return NewFromTokens(toks), nil
}

func NewFromTokens(toks []token.Token) *Parser {
	p := &Parser{idx: -1, Toks: toks}

	p.next()

	return p
}

func (p *Parser) Parse() (node.Statements, error) {
	var statements node.Statements

	for statement, err := p.NextStatement(); true; statement, err = p.NextStatement() {
		if err != nil {
			return statements, err
		}

		if statement == nil {
			break
		}

		statements.List = append(statements.List, statement)
	}

	return statements, nil
}

func (p *Parser) NextStatement() (node.Statement, error) {
	for {
		switch p.tok.Type {
		case token.EOF: return nil, nil

		case token.Separator:
			p.next()

			continue

		default: return p.parseStatement()
		}
	}
}

func (p *Parser) parseStatement() (node.Statement, error) {
	return p.parseLogicalBinOp()
}

func (p *Parser) parseLogicalBinOp() (node.Statement, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	// If there are no logical operators, just return the parsed node
	if isStatementEnd(p.tok) {
		return left, nil
	}

	// Else parse the logical operators
	for p.tok.Type == token.And || p.tok.Type == token.Or {
		tok := *p.tok // Save the operator token for the operator node
		p.next()

		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		left = &node.BinOpStatement{Token: tok, Left: left, Right: right}

		if isStatementEnd(p.tok) {
			break
		}
	}

	return left, nil
}

func (p *Parser) parseFactor() (node.Statement, error) {
	switch p.tok.Type {
	case token.String: return p.parseCmd()
	case token.Exit:   return p.parseExit()
	case token.Echo:   return p.parseEcho()
	case token.Cd:     return p.parseCd()

	case token.LParen:
		p.next()

		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		} else if p.tok.Type != token.RParen {
			return nil, errors.ExpectedToken(p.tok, token.RParen)
		}

		p.next()

		return s, nil

	default: return nil, errors.UnexpectedToken(p.tok)
	}
}

func (p *Parser) parseCmd() (*node.CmdStatement, error) {
	cs := &node.CmdStatement{Token: *p.tok, Cmd: p.tok.Data}

	// Get the command arguments
	for p.next(); !isArgsEnd(p.tok); p.next() {
		var arg string

		switch p.tok.Type {
		// Arguments can only be strings or integers
		case token.String, token.Integer: arg = p.tok.Data

		default: return nil, errors.UnexpectedToken(p.tok)
		}

		cs.Args = append(cs.Args, arg)
	}

	return cs, nil
}

func (p *Parser) parseExit() (*node.ExitStatement, error) {
	es := &node.ExitStatement{Token: *p.tok}

	// Exitcodes have to be integers

	// If there is just an exit token alone, we will use the latest exitcode to exit
	if p.next(); isArgsEnd(p.tok) {
		es.HasEx = false // The statement has no requested exitcode
	} else if p.tok.Type == token.Integer {
		// If there is a requested exitcode, save it
		es.HasEx = true

		ex, err := strconv.Atoi(p.tok.Data)
		if err != nil {
			panic(err)
		}

		es.Ex = ex

		p.next()
	} else {
		return nil, errors.ExpectedToken(p.tok, token.Integer)
	}

	// Make sure the statement is ended
	if !isArgsEnd(p.tok) {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return es, nil
}

func (p *Parser) parseEcho() (*node.EchoStatement, error) {
	echo := &node.EchoStatement{Token: *p.tok}

	// Append all the arguments to a string
	for p.next(); !isArgsEnd(p.tok); p.next() {
		switch p.tok.Type {
		case token.String, token.Integer: echo.Msg += p.tok.Data + " "

		default: return nil, errors.UnexpectedToken(p.tok)
		}
	}

	return echo, nil
}

func (p *Parser) parseCd() (*node.CdStatement, error) {
	cd := &node.CdStatement{Token: *p.tok}

	// If no path is specified, we change the directory to the home directory
	if p.next(); isArgsEnd(p.tok) {
		cd.Path = "~/"
	} else if p.tok.Type == token.String {
		cd.Path = p.tok.Data

		p.next()
	} else {
		return nil, errors.ExpectedToken(p.tok, token.String)
	}

	return cd, nil
}

func (p *Parser) next() {
	// Make sure to not run over the source end
	if p.idx + 1 < len(p.Toks) {
		p.idx ++
		p.tok = &p.Toks[p.idx]
	}
}

func (p *Parser) peekTok() token.Token {
	if p.tok.Type == token.EOF {
		return *p.tok
	} else {
		return p.Toks[p.idx + 1]
	}
}
