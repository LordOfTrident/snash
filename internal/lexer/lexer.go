package lexer

import (
	"unicode"

	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/token"
)

type Lexer struct {
	where token.Where

	idx  int
	char rune

	source string
}

func New(source, path string) *Lexer {
	l := &Lexer{where: token.Where{Row: 1, Path: path}, idx: -1, source: source}
	l.next()

	return l
}

func (l *Lexer) Lex() ([]token.Token, error)  {
	toks := []token.Token{}

	end := false
	for !end {
		tok := l.NextToken()
		switch tok.Type {
		case token.EOF:   end = true
		case token.Error: return toks, errors.ErrorTokenToError(tok)
		}

		toks = append(toks, tok)
	}

	return toks, nil
}

func (l *Lexer) NextToken() (tok token.Token) {
	for {
		switch l.char {
		// EOF token marks the source end
		case '\x00': tok = token.NewEOF(l.where)

		case '\n', ';':
			tok = token.New(token.Separator, "", l.where, 1)
			l.next()

		case '=':
			tok = token.New(token.Equals, string(l.char), l.where, 1)
			l.next()

		// Ignore whitespaces
		case ' ', '\r', '\t', '\v', '\f':
			l.next()

			continue

		case '#':
			l.skipComment()

			continue

		case '&': tok = l.lexAnd()
		case '|': tok = l.lexOr()

		default:
			if unicode.IsDigit(l.char) {
				tok = l.lexInteger()
			} else {
				tok = l.lexWord()
			}
		}

		break
	}

	return
}

func (l *Lexer) skipComment() {
	for l.char != '\x00' && l.char != '\n' {
		l.next()
	}
}

func (l *Lexer) lexAnd() token.Token {
	start := l.where

	if l.next(); l.char != '&' {
		return token.NewError(start, 1,
		                      "Unexpected character \"&\", did you mean \"&&\"?")
	}

	l.next()

	return token.New(token.And, "&&", start, 2)
}

func (l *Lexer) lexOr() token.Token {
	start := l.where

	// TODO: Add piping
	if l.next(); l.char != '|' {
		return token.NewError(start, 1,
		                      "Unexpected character \"|\", did you mean \"||\"?")
	}

	l.next()

	return token.New(token.Or, "||", start, 2)
}

func (l *Lexer) lexWord() token.Token {
	start := l.where // The starting position of the token
	str   := ""      // The token data string

	apostrophe := '\x00' // To save the current apostrophe we are using
	escape     := false  // Are we inside an escape sequence?

	isBareWord := true // Could be a keyword

loop:
	for ; apostrophe != '\x00' || !unicode.IsSpace(l.char) && l.char != ';'; l.next() {
		switch l.char {
		case '\x00':
			if apostrophe == '\x00' {
				break loop
			} else {
				return token.NewError(start, l.where.Col - start.Col, "String not terminated")
			}

		// Whitespaces and other special characters are allowed inside apostrophes
		case '\'', '"', '`':
			if escape {
				// If we are escaping the apostrophe, add it to the string
				str += string(l.char)

				escape = false
			} else {
				// Find out if it is an end marking apostrophe, a beginning one or just a part
				// of the string
				if apostrophe == l.char {
					apostrophe = '\x00'
				} else if apostrophe == '\x00' {
					apostrophe = l.char
				} else {
					str += string(l.char)
				}
			}

		case '\\':
			if apostrophe != '"' && apostrophe != '`' { // Escape sequences are only allowed
			                                            // inside of " and ` apostrophes
				str += string(l.char)
			} else if escape {
				str += string(l.char)

				escape = false
			} else {
				escape = true
			}

		case '\n':
			// Multi line strings
			if apostrophe == '`' {
				str += string(l.char)
			} else {
				return token.NewError(start, l.where.Col - start.Col, "String exceeds line")
			}

		case '$':
			if escape || apostrophe == '\'' {
				str += "\\$"

				escape = false
			} else {
				str += "$"
			}

		default:
			if escape {
				// Parse the escape sequence
				switch l.char {
				case 'e': str += string(27)
				case 'n': str += string('\n')
				case 'r': str += string('\r')
				case 't': str += string('\t')
				case 'v': str += string('\v')
				case 'b': str += string('\b')
				case 'f': str += string('\f')

				default:
					return token.NewError(start, l.where.Col - start.Col,
					                      "Unknown escape sequence \"\\%c\"", l.char)
				}

				escape = false
			} else {
				str += string(l.char)
			}
		}

		if isBareWord {
			isBareWord = unicode.IsLetter(l.char)
		}
	}

	// Check if the string is a keyword
	if isBareWord {
		return token.New(getBareWordTokenType(str), str, start, l.where.Col - start.Col)
	} else {
		return token.New(token.Word, str, start, l.where.Col - start.Col)
	}
}

func getBareWordTokenType(word string) token.Type {
	switch word {
	case "help": return token.Help
	case "exit": return token.Exit
	case "echo": return token.Echo
	case "cd":   return token.Cd

	case "let":    return token.Let
	case "export": return token.Export

	default: return token.BareWord
	}
}

func (l *Lexer) next() {
	l.idx ++

	// Make sure we wont exceed the source code length
	if l.idx >= len(l.source) {
		l.char = '\x00'
	} else {
		l.char = rune(l.source[l.idx])
	}

	// Update position variables
	if l.char == '\n' {
		l.where.Col = 0
		l.where.Row ++
	} else {
		l.where.Col ++
	}
}

func (l *Lexer) peekChar() rune {
	if l.idx + 1 >= len(l.source) {
		return '\x00'
	} else {
		return rune(l.source[l.idx + 1])
	}
}

func (l *Lexer) lexInteger() token.Token {
	start := l.where // Save the token starting position
	str   := ""      // The token data string

	// TODO: make tokens like '123abc' not error and instead be lexer as strings
	for ; l.char != '\x00' && !unicode.IsSpace(l.char) && l.char != ';'; l.next() {
		if !unicode.IsDigit(l.char) {
			return token.NewError(start, l.where.Col - start.Col,
			                      "Unexpected character \"%c\" in number", l.char)
		}

		str += string(l.char)
	}

	return token.New(token.Integer, str, start, l.where.Col - start.Col)
}
