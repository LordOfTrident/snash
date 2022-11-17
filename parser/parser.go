package parser

import (
	"strconv"

	"github.com/LordOfTrident/snash/err"
	"github.com/LordOfTrident/snash/token"
	"github.com/LordOfTrident/snash/lexer"
	"github.com/LordOfTrident/snash/node"
)

func unexpected(tok *token.Token) error {
	return err.New(tok.Where, "Unexpected %v", tok)
}

func expected(tok *token.Token, expected token.Type) error {
	return err.New(tok.Where, "Expected type '%v', got %v", expected, tok)
}

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
	p := &Parser{idx: -1}

	var err error
	p.Toks, err = l.Lex()
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
		var err       error // Save the error to handle all errors at once at the back of the loop

		switch p.tok.Type {
		case token.String:      statement, err = p.parseCmd()
		case token.KeywordExit: statement, err = p.parseExit()
		case token.KeywordEcho: statement, err = p.parseEcho()
		case token.KeywordCd:   statement, err = p.parseCd()

		case token.Separator:
			p.next()

			continue

		default: err = unexpected(p.tok)
		}

		if err != nil {
			return program, err
		}

		program.List = append(program.List, statement)
	}

	return program, nil
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

		default: return nil, unexpected(p.tok)
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
		return nil, expected(p.tok, token.Integer)
	}

	// Make sure the statement is ended
	if !isStatementEnd(p.tok) {
		return nil, expected(p.tok, token.Separator)
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

		default: return nil, unexpected(p.tok)
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
		return nil, expected(p.tok, token.String)
	}

	return cd, nil
}
