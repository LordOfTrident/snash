package lexer

import (
	"fmt"

	"github.com/LordOfTrident/snash/pkg/token"
)

const CharNone byte = 0

func IsWhitespace(char byte) bool {
	switch char {
	case ' ', '\r', '\t', '\v', '\f': return true

	default: return false
	}
}

func IsSeparator(char byte) bool {
	switch char {
	case ' ', '\r', '\t', '\n', '\v', '\f', ';': return true

	default: return false
	}
}

func IsAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func IsDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

type Lexer struct {
	where token.Where

	idx  int
	char byte

	source string
}

func New(source, path string) *Lexer {
	l := &Lexer{where: token.Where{Row: 1, Path: path}, idx: -1, source: source}
	l.next()

	return l
}

func (l *Lexer) Lex() ([]token.Token, error)  {
	toks := []token.Token{}

	for l.char != CharNone {
		var tok token.Token

		switch l.char {
		case '\n', ';':
			tok = token.New(token.Separator, "", l.where, 1)
			l.next()

		// Ignore whitespaces
		case ' ', '\r', '\t', '\v', '\f':
			l.next()

			continue

		case '#':
			l.skipComment()

			continue

		default:
			if IsDigit(l.char) {
				tok = l.lexInteger()
			} else {
				tok = l.lexString()
			}
		}

		if tok.Type == token.Error {
			return nil, fmt.Errorf("%v: %v", tok.Where, tok.Data)
		}

		toks = append(toks, tok)
	}

	// Add an EOF token to mark the source end
	toks = append(toks, token.NewEOF(l.where))

	return toks, nil
}

func (l *Lexer) next() {
	l.idx ++

	// Make sure we wont exceed the source code length
	if l.idx >= len(l.source) {
		l.char = CharNone
	} else {
		l.char = l.source[l.idx]
	}

	// Update position variables
	if l.char == '\n' {
		l.where.Col = 0
		l.where.Row ++
	} else {
		l.where.Col ++
	}
}

func (l *Lexer) peekChar() byte {
	if l.idx + 1 >= len(l.source) {
		return CharNone
	} else {
		return l.source[l.idx + 1]
	}
}

func (l *Lexer) skipComment() {
	for l.char != CharNone && l.char != '\n' {
		l.next()
	}
}

func (l *Lexer) lexString() token.Token {
	start := l.where // The starting position of the token
	str   := ""      // The token data string

	apostrophe := CharNone // To save the current apostrophe we are using
	escape     := false    // Are we inside an escape sequence?

	isWord := true // A flag to see if the string could be a keyword

	// TODO: finish this function
	for ; l.char != CharNone && (apostrophe != CharNone || !IsSeparator(l.char)); l.next() {
		if isWord {
			isWord = IsAlpha(l.char)
		}

		switch l.char {
		// Whitespaces and other special characters are allowed inside apostrophes
		case '\'', '"':
			if escape {
				// If we are escaping the apostrophe, add it to the string
				str += string(l.char)

				escape = false
			} else {
				// Find out if it is an end marking apostrophe, a beginning one or just a part
				// of the string
				if apostrophe == l.char {
					apostrophe = CharNone
				} else if apostrophe == CharNone {
					apostrophe = l.char
				} else {
					str += string(l.char)
				}
			}

		case '\\':
			if escape || apostrophe == CharNone { // Escape sequences are not allowed
			                                      // outside of apostrophes
				str += string(l.char)
			} else if escape {
				str += string(l.char)

				escape = false
			} else {
				escape = true
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
				case '0': str += string(0)

				default:
					return token.NewError(l.where, "Unknown escape sequence '\\%c'", l.char)
				}

				escape = false
			} else {
				str += string(l.char)
			}
		}
	}

	if apostrophe != CharNone {
		return token.NewError(l.where, "String exceeds line")
	}

	// Check if the string is a keyword
	if isWord {
		return token.New(getWordTokenType(str), str, start, l.where.Col - start.Col)
	} else {
		return token.New(token.String, str, start, l.where.Col - start.Col)
	}
}

func getWordTokenType(word string) token.Type {
	switch word {
	case "exit": return token.KeywordExit
	case "echo": return token.KeywordEcho
	case "cd":   return token.KeywordCd

	default: return token.String
	}
}

func (l *Lexer) lexInteger() token.Token {
	start := l.where // Save the token starting position
	str   := ""      // The token data string

	// TODO: make tokens like '123abc' not error and instead be lexer as strings
	for ; l.char != CharNone && !IsSeparator(l.char); l.next() {
		if !IsDigit(l.char) {
			return token.NewError(l.where, "Unexpected character '%c' in number", l.char)
		}

		str += string(l.char)
	}

	return token.New(token.Integer, str, start, l.where.Col - start.Col)
}
