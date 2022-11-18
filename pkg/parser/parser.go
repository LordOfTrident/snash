package parser

import (
	"strconv"

	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/lexer"
	"github.com/LordOfTrident/snash/pkg/node"
)

func isStatementEnd(tok *token.Token) bool {
	return tok.Type == token.Separator || tok.Type == token.EOF
}

type Parser struct {
	idx int
	tok *token.Token

	Toks []token.Token
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
	var program node.Statements

	for statement, err := p.NextStatement(); true; statement, err = p.NextStatement() {
		if err != nil {
			return program, err
		}

		if statement == nil {
			break
		}

		program.List = append(program.List, statement)
	}

	return program, nil
}

func (p *Parser) NextStatement() (node.Statement, error) {
	for {
		switch p.tok.Type {
		case token.EOF: return nil, nil

		case token.String:      return p.parseCmd()
		case token.KeywordExit: return p.parseExit()
		case token.KeywordEcho: return p.parseEcho()
		case token.KeywordCd:   return p.parseCd()

		case token.Separator:
			p.next()

			continue

		default: return nil, errors.UnexpectedToken(p.tok)
		}
	}
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

func (p *Parser) parseCmd() (*node.CmdStatement, error) {
	cs := &node.CmdStatement{Token: *p.tok, Cmd: p.tok.Data}

	// Get the command arguments
	for p.next(); !isStatementEnd(p.tok); p.next() {
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
	if p.next(); isStatementEnd(p.tok) {
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
	if !isStatementEnd(p.tok) {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return es, nil
}

func (p *Parser) parseEcho() (*node.EchoStatement, error) {
	echo := &node.EchoStatement{Token: *p.tok}

	// Append all the arguments to a string
	for p.next(); !isStatementEnd(p.tok); p.next() {
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
	if p.next(); isStatementEnd(p.tok) {
		cd.Path = "~/"
	} else if p.tok.Type == token.String {
		cd.Path = p.tok.Data

		p.next()
	} else {
		return nil, errors.ExpectedToken(p.tok, token.String)
	}

	return cd, nil
}
