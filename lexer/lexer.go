package lexer

import (
	"github.com/LordOfTrident/snash/compilerError"
	"github.com/LordOfTrident/snash/token"
)

const charNone byte = 0

func isSeparator(char byte) bool {
	switch char {
	case ' ', '\r', '\t', '\n', '\v', '\f', ';': return true;

	default: return false;
	}
}

func isAlpha(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isDigit(char byte) bool {
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

	for l.char != charNone {
		var tok token.Token
		var err error

		switch l.char {
		case '\n', ';':
			tok = token.New(token.Separator, "", l.where)
			l.next()

		case ' ', '\r', '\t', '\v', '\f':
			l.next()

			continue

		case '#':
			l.skipComment()

			continue

		default:
			if isDigit(l.char) {
				tok, err = l.lexInteger()
			} else {
				tok, err = l.lexString()
			}
		}

		if err != nil {
			return nil, err
		}

		toks = append(toks, tok)
	}

	toks = append(toks, token.New(token.EOF, "", l.where))

	return toks, nil
}

func (l *Lexer) next() {
	l.idx ++

	if l.idx >= len(l.source) {
		l.char = charNone
	} else {
		l.char = l.source[l.idx]
	}

	if l.char == '\n' {
		l.where.Col = 0
		l.where.Row ++
	} else {
		l.where.Col ++
	}
}

func (l *Lexer) peekChar() byte {
	if l.idx + 1 >= len(l.source) {
		return charNone
	} else {
		return l.source[l.idx + 1]
	}
}

func (l *Lexer) skipComment() {
	for l.char != charNone && l.char != '\n' {
		l.next()
	}
}

func (l *Lexer) lexString() (token.Token, error) {
	start := l.where
	str   := ""

	apostrophe := charNone
	escape     := false

	isWord := true

	// TODO: finish this function
	for ; l.char != charNone && (apostrophe != charNone || !isSeparator(l.char)); l.next() {
		if isWord {
			isWord = isAlpha(l.char)
		}

		switch l.char {
		case '\'', '"':
			if escape {
				str += string(l.char)

				escape = false
			} else {
				if apostrophe == l.char {
					apostrophe = charNone
				} else if apostrophe == charNone {
					apostrophe = l.char
				} else {
					str += string(l.char)
				}
			}

		case '\\':
			if escape || apostrophe == charNone {
				str += string(l.char)
			} else if escape {
				str += string(l.char)

				escape = false
			} else {
				escape = true
			}

		default:
			if escape {
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
					return token.Empty(),
					       compilerError.New(l.where, "Unknown escape sequence '\\%c'", l.char)
				}

				escape = false
			} else {
				str += string(l.char)
			}
		}
	}

	if apostrophe != charNone {
		return token.Empty(), compilerError.New(l.where, "String exceeds line")
	}

	if isWord {
		return token.New(getWordTokenType(str), str, start), nil
	} else {
		return token.New(token.String, str, start), nil
	}
}

func getWordTokenType(word string) token.Type {
	switch word {
	case "exit": return token.KeywordExit

	default: return token.String
	}
}

func (l *Lexer) lexInteger() (token.Token, error) {
	start := l.where
	str   := ""

	for ; l.char != charNone && !isSeparator(l.char); l.next() {
		if !isDigit(l.char) {
			return token.Empty(),
			       compilerError.New(l.where, "Unexpected character '%c' in number", l.char)
		}

		str += string(l.char)
	}

	return token.New(token.Integer, str, start), nil
}
