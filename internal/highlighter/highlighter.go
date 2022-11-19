package highlighter

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/attr"
	"github.com/LordOfTrident/snash/internal/token"
	"github.com/LordOfTrident/snash/internal/lexer"
)

const (
	colorError   = attr.Underline + attr.BrightRed
	colorComment = attr.Italics   + attr.Grey
	colorKeyword = attr.Bold      + attr.BrightBlue
	colorCmd     = attr.Bold      + attr.BrightYellow
	colorInteger = attr.BrightCyan
	colorPath    = attr.Underline + attr.BrightGreen
	colorEscape  = attr.BrightMagenta
	colorString  = attr.BrightGreen
)

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%v%vError:%v %v\n", attr.Bold, attr.BrightRed, attr.Reset,
	            HighlightStrings(err.Error()))
}

func Printf(format string, args... interface{}) {
	fmt.Print(HighlightStrings(fmt.Sprintf(format, args...)))
}

func cmdExists(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func isCmd(toks []token.Token, i int) (isCmd bool) {
	// Would the token be parsed as a command?
	isCmd = true
	if i > 0 {
		if toks[i - 1].Type != token.Separator {
			isCmd = false
		}
	}

	return
}

func HighlightStrings(str string) (ret string) {
	apostrophe := utils.CharNone
	escape     := false

	for _, ch := range str {
		ch := byte(ch)

		switch ch {
		case '\'', '"', '`':
			if escape {
				escape = false
				ret += string(ch) + attr.Reset + colorString

				continue
			} else {
				// Find out if it is an end marking apostrophe, a beginning one or just a part
				// of the string
				if apostrophe == ch {
					apostrophe = utils.CharNone

					ret += string(ch) + attr.Reset

					continue // We already added the character to the string
				} else if apostrophe == utils.CharNone {
					apostrophe = ch
					ret += colorString
				}
			}

		case '\\':
			if apostrophe == '"' || apostrophe == '`' { // Escape sequences are only allowed
				if escape {                             // inside of " and ` apostrophes
					escape = false

					ret += string(ch) + attr.Reset + colorString

					continue
				} else {
					escape = true
					ret += colorEscape
				}
			}

		default:
			if escape {
				ret += string(ch) + attr.Reset + colorString

				continue
			}
		}

		ret += string(ch)
	}

	return
}

func highlightNext(toks []token.Token, i int, line string) (highlighted string, err error) {
	tok := toks[i]
	col := tok.Where.Col - 1

	prevCol := 0
	if i > 0 {
		prevCol = (toks[i - 1].Where.Col - 1 + toks[i - 1].TxtLen)
	}

	// If there is a space between this and the previous token
	if col - prevCol > 0 {
		// Save the ignored characters in between tokens and color the comments
		highlighted += strings.Replace(line[prevCol:col], "#", colorComment + "#", -1)
	}

	if tok.Type != token.EOF {
		// Get the raw token text
		txt := line[col:col + tok.TxtLen]

		isCmd := isCmd(toks, i)

		switch tok.Type {
		case token.Integer: highlighted += colorKeyword + txt
		case token.Error:   highlighted += colorError   + txt

		default:
			if tok.Type.IsKeyword() {
				highlighted += colorKeyword + txt
			} else if isCmd { // Is the current token a command?
				if cmdExists(tok.Data) {
					highlighted += colorCmd + txt
				} else {
					err = errors.CmdNotFound(tok.Data, tok.Where)

					highlighted += colorError + txt
				}
			} else if !isCmd && fileExists(tok.Data) { // Is the current token a file path argument?
				highlighted += colorPath + txt
			} else {
				highlighted += HighlightStrings(txt)
			}
		}
	}

	highlighted += attr.Reset

	return
}

func HighlightLine(line, path string) (out string, firstErr error) {
	l := lexer.New(line, path)

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

		next, err := highlightNext(toks, i, line)
		if err != nil && firstErr == nil {
			firstErr = err
		}

		out += next
	}

	out += attr.Reset

	return
}
