package parser

import (
	"strconv"

	"github.com/LordOfTrident/snash/compilerError"
	"github.com/LordOfTrident/snash/token"
	"github.com/LordOfTrident/snash/lexer"
	"github.com/LordOfTrident/snash/node"
)

type Parser struct {
	idx int
	tok *token.Token

	toks []token.Token
}

func New(source, path string) (*Parser, error) {
	l := lexer.New(source, path)
	p := &Parser{idx: -1}

	var err error
	p.toks, err = l.Lex()
	if err != nil {
		return nil, err
	}

	p.next()

	return p, nil
}

func (p *Parser) Parse() (node.Statements, error) {
	var program node.Statements

	for p.tok.Type != token.EOF {
		var statement node.Statement
		var err       error

		switch p.tok.Type {
		case token.String:      statement, err = p.parseCmd()
		case token.KeywordExit: statement, err = p.parseExit()

		case token.Separator:
			p.next()

			continue

		default: err = compilerError.UnexpectedToken(p.tok)
		}

		if err != nil {
			return program, err
		}

		program.List = append(program.List, statement)
	}

	return program, nil
}

func (p *Parser) next() {
	p.idx ++
	p.tok = &p.toks[p.idx]
}

func (p *Parser) peekTok() token.Token {
	if p.tok.Type == token.EOF {
		return *p.tok
	} else {
		return p.toks[p.idx + 1]
	}
}

func (p *Parser) parseCmd() (*node.CmdStatement, error) {
	cs := &node.CmdStatement{Token: *p.tok, Cmd: p.tok.Data}

	for p.next(); !p.tok.IsStatementEnd(); p.next() {
		var arg string

		switch p.tok.Type {
		case token.String, token.Integer: arg = p.tok.Data

		default: return nil, compilerError.UnexpectedToken(p.tok)
		}

		cs.Args = append(cs.Args, arg)
	}

	return cs, nil
}

func (p *Parser) parseExit() (*node.ExitStatement, error) {
	es := &node.ExitStatement{Token: *p.tok}

	if p.next(); p.tok.Type != token.Integer {
		return nil, compilerError.ExpectedToken(p.tok, token.Integer)
	}

	ex, err := strconv.Atoi(p.tok.Data)
	if err != nil {
		panic(err)
	}

	es.Exitcode = ex

	if p.next(); p.tok.Type != token.Separator {
		return nil, compilerError.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return es, nil
}
