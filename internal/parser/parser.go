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
	if p.tok.IsStatementEnd() {
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

		if p.tok.IsStatementEnd() {
			break
		}
	}

	return left, nil
}

func (p *Parser) parseFactor() (node.Statement, error) {
	switch p.tok.Type {
	case token.Word, token.BareWord:
		if p.peekTok().Type == token.Equals {
			return p.parseAssign()
		} else {
			return p.parseCmd()
		}

	case token.Let:    return p.parseLet()
	case token.Export: return p.parseExport()

	case token.Help: return p.parseHelp()
	case token.Exit: return p.parseExit()
	case token.Echo: return p.parseEcho()
	case token.Cd:   return p.parseCd()

	default: return nil, errors.UnexpectedToken(p.tok)
	}
}

func (p *Parser) parseExport() (*node.ExportStatement, error) {
	export := &node.ExportStatement{Token: *p.tok}

	// Variable to export
	if p.next(); !p.tok.IsString() {
		return nil, errors.ExpectedToken(p.tok, token.Word)
	} else {
		export.Name = p.tok.Data
	}

	if p.next(); !p.tok.IsArgsEnd() {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return export, nil
}

func (p *Parser) parseLet() (*node.LetStatement, error) {
	let := &node.LetStatement{Token: *p.tok}

	// Variable identifier
	if p.next(); !p.tok.IsString() {
		return nil, errors.ExpectedToken(p.tok, token.Word)
	} else {
		let.Name = p.tok.Data
	}

	if p.next(); p.tok.Type != token.Equals {
		return nil, errors.ExpectedToken(p.tok, token.Equals)
	}

	// Variable value
	if p.next(); !p.tok.IsString() {
		return nil, errors.ExpectedToken(p.tok, token.Word)
	} else {
		let.Value = p.tok.Data
	}

	if p.next(); !p.tok.IsArgsEnd() {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return let, nil
}

func (p *Parser) parseAssign() (*node.AssignStatement, error) {
	as := &node.AssignStatement{Token: *p.tok, Name: p.tok.Data}

	if p.next(); p.tok.Type != token.Equals {
		return nil, errors.ExpectedToken(p.tok, token.Equals)
	}

	// New variable value
	if p.next(); !p.tok.IsString() {
		return nil, errors.ExpectedToken(p.tok, token.Word)
	} else {
		as.Value = p.tok.Data
	}

	if p.next(); !p.tok.IsArgsEnd() {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return as, nil
}

func (p *Parser) parseCmd() (*node.CmdStatement, error) {
	cs := &node.CmdStatement{Token: *p.tok, Cmd: p.tok.Data}

	// Get the command arguments
	for p.next(); !p.tok.IsArgsEnd(); p.next() {
		var arg string

		if p.tok.IsArg() {
			arg = p.tok.Data
		} else {
			return nil, errors.UnexpectedToken(p.tok)
		}

		cs.Args = append(cs.Args, arg)
	}

	return cs, nil
}

func (p *Parser) parseHelp() (*node.HelpStatement, error) {
	hs := &node.HelpStatement{Token: *p.tok}

	// 'help' takes no arguments
	if p.next(); !p.tok.IsArgsEnd() {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return hs, nil
}

func (p *Parser) parseExit() (*node.ExitStatement, error) {
	es := &node.ExitStatement{Token: *p.tok}

	// Exitcodes have to be integers

	// If there is just an exit token alone, we will use the latest exitcode to exit
	if p.next(); p.tok.IsArgsEnd() {
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
	if !p.tok.IsArgsEnd() {
		return nil, errors.ExpectedToken(p.tok, token.Separator)
	}

	p.next()

	return es, nil
}

func (p *Parser) parseEcho() (*node.EchoStatement, error) {
	echo := &node.EchoStatement{Token: *p.tok}

	// Append all the arguments to a string
	for p.next(); !p.tok.IsArgsEnd(); p.next() {
		if p.tok.IsArg() {
			echo.Msg += p.tok.Data + " "
		} else {
			return nil, errors.UnexpectedToken(p.tok)
		}
	}

	return echo, nil
}

func (p *Parser) parseCd() (*node.CdStatement, error) {
	cd := &node.CdStatement{Token: *p.tok}

	// If no path is specified, we change the directory to the home directory
	if p.next(); p.tok.IsArgsEnd() {
		cd.Path = "~/"
	} else if p.tok.IsString() {
		cd.Path = p.tok.Data

		p.next()
	} else {
		return nil, errors.ExpectedToken(p.tok, token.Word)
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
