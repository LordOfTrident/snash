package highlighter

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/LordOfTrident/snash/pkg/term"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/token"
	"github.com/LordOfTrident/snash/internal/lexer"
	"github.com/LordOfTrident/snash/internal/env"
)

const (
	colorError   = term.AttrUnderline + term.AttrBrightRed
	colorComment = term.AttrItalics   + term.AttrGrey

	colorKeyword  = term.AttrBold + term.AttrBrightBlue
	colorOperator = term.AttrBold + term.AttrMagenta

	colorCmd  = term.AttrBold      + term.AttrBrightYellow
	colorPath = term.AttrUnderline + term.AttrBrightGreen

	colorInteger = term.AttrBrightCyan

	colorString = term.AttrBrightGreen
	colorEscape = term.AttrBrightMagenta
	colorVar    = term.AttrBrightCyan
)

type Highlighter struct {
	env *env.Env
}

func New(env *env.Env) *Highlighter {
	return &Highlighter{env: env}
}

func (h *Highlighter) Highlight(code, path string) (out string, firstErr error) {
	l := lexer.New(code, path)

	// Lex all the tokens (We do this manually because we want to keep lexing even after errors)
	var toks []token.Token
	for tok := l.NextToken(); true; tok = l.NextToken() {
		toks = append(toks, tok)

		if tok.Type == token.EOF {
			break
		}
	}

	for i, tok := range toks {
		// If an error was found, only report it if it is the first error
		if tok.Type == token.Error && firstErr == nil {
			firstErr = errors.ErrorTokenToError(tok)
		}

		next, err := highlightNext(toks, i, code)
		if err != nil && firstErr == nil {
			firstErr = err
		}

		out += next
	}

	out += term.AttrReset

	return
}

func cmdExists(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

func isCmd(toks []token.Token, i int) (isCmd bool) {
	// Would the token be parsed as a command?
	isCmd = true
	if i > 0 {
		if !toks[i - 1].IsArgsEnd() {
			isCmd = false
		}
	}

	// Variable assignments are not command calls
	if i + 1 < len(toks) {
		if toks[i + 1].Type == token.Equals {
			isCmd = false
		}
	}

	if !toks[i].IsString() {
		return false
	}

	return
}

func PrintError(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, "%v%vError:%v %v\n", term.AttrBold, term.AttrBrightRed, term.AttrReset,
	            HighlightStrings(fmt.Sprintf(format, args...)))
}

func Printf(format string, args... interface{}) {
	fmt.Print(HighlightStrings(fmt.Sprintf(format, args...)))
}

func HighlightStrings(str string) (ret string) {
	apostrophe := '\x00'
	escape     := false
	inVar      := false

	for _, ch := range str {
		if inVar {
			// Only alphanum and _ characters are allowed in variable names
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
				ret += term.AttrReset
				if apostrophe != '\x00' {
					ret += colorString
				}

				inVar = false
			}
		} else {
			switch ch {
			case '\'', '"', '`':
				if escape {
					escape = false
					ret += string(ch) + term.AttrReset + colorString

					continue
				} else {
					// Find out if it is an end marking apostrophe, a beginning one or just a part
					// of the string
					if apostrophe == ch {
						apostrophe = '\x00'

						ret += string(ch) + term.AttrReset

						continue // We already added the character to the string
					} else if apostrophe == '\x00' {
						apostrophe = ch
						ret += colorString
					}
				}

			case '\\':
				if apostrophe == '"' || apostrophe == '`' { // Escape sequences are only allowed
					if escape {                             // inside of " and ` apostrophes
						escape = false

						ret += string(ch) + term.AttrReset + colorString

						continue
					} else {
						escape = true
						ret += colorEscape
					}
				}

			case '$':
				if escape || apostrophe == '\'' {
					ret += "$" + term.AttrReset + colorString

					escape = false
				} else {
					ret += colorVar + "$"

					inVar = true
				}

				continue

			default:
				if escape {
					ret += string(ch) + term.AttrReset + colorString

					escape = false

					continue
				}
			}
		}

		ret += string(ch)
	}

	return
}

func highlightNext(toks []token.Token, i int, code string) (highlighted string, err error) {
	tok := toks[i]
	col := tok.Where.Col - 1

	prevCol := 0
	if i > 0 {
		prevCol = (toks[i - 1].Where.Col - 1 + toks[i - 1].TxtLen)
	}

	// If there is a space between this and the previous token
	if col - prevCol > 0 {
		// Save the ignored characters in between tokens and color the comments
		highlighted += strings.Replace(code[prevCol:col], "#", colorComment + "#", -1)
	}

	if tok.Type != token.EOF {
		// Get the raw token text
		txt := code[col:col + tok.TxtLen]

		isCmd := isCmd(toks, i)

		switch tok.Type {
		case token.Integer: highlighted += colorKeyword + txt
		case token.Error:   highlighted += colorError   + txt

		default:
			if tok.IsKeyword() {
				highlighted += colorKeyword + txt
			} else if tok.IsOp() {
				highlighted += colorOperator + txt
			} else if isCmd { // Is the current token a command?
				if cmdExists(tok.Data) {
					highlighted += colorCmd + txt
				} else {
					err = errors.CmdNotFound(tok.Data, tok.Where)

					highlighted += colorError + txt
				}
			} else if !isCmd && utils.FileExists(tok.Data) { // Is the current token a
			                                                 // file path argument?
				highlighted += colorPath + txt
			} else {
				highlighted += HighlightStrings(txt)
			}
		}
	}

	highlighted += term.AttrReset

	return
}
