package highlighter

import (
	"os"
	"os/exec"
	"strings"

	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/attr"
	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/parser"
)

const (
	colorError   = attr.Underline + attr.BrightRed
	colorComment = attr.Italics   + attr.Grey
	colorKeyword = attr.Bold      + attr.BrightBlue
	colorCmd     = attr.Bold      + attr.BrightYellow
	colorInteger = attr.BrightCyan
	colorPath    = attr.BrightGreen
)

func cmdExists(name string) bool {
	_, err := exec.LookPath(name)

	return err == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func highlightNext(tok token.Token, line string, lastCol int, isCmd bool) (int, string, error) {
	idx := tok.Where.Col - 1

	// Save the trailing spaces to output them
	var spaces, highlighted string
	if idx - lastCol > 0 {
		spaces = line[lastCol:idx]
	}

	if tok.Type == token.EOF {
		highlighted += spaces
	} else {
		// Get the raw token text
		txt := line[idx:idx + tok.TxtLen]

		// Update the last column
		lastCol = idx + tok.TxtLen

		switch tok.Type {
		case token.Integer: highlighted += spaces + colorInteger + txt

		default:
			if tok.Type.IsKeyword() {
				highlighted += spaces + colorKeyword + txt
			} else if isCmd { // Is the current token a command?
				if cmdExists(tok.Data) {
					highlighted += spaces + colorCmd + txt
				} else {
					return lastCol, spaces + colorError + txt + attr.Reset,
					       errors.CmdNotFound(tok.Data, tok.Where)
				}
			} else if !isCmd && fileExists(tok.Data) { // Is the current token a file path argument?
				highlighted += spaces + colorPath + txt
			} else {
				highlighted += spaces + txt
			}
		}

		// Reset the color at the end of the token
		highlighted += attr.Reset
	}

	return lastCol, highlighted, nil
}

// TODO: improve how the highlighting works, maybe lex the tokens one by one as the highlighting
//       loop goes

func HighlightLine(line, path string) (string, error) {
	// Try parsing and lexing the code to catch any errors
	p, err := parser.New(line, path)
	if err != nil {
		return colorError + line + attr.Reset, err
	}

	_, err = p.Parse()
	if err != nil {
		return colorError + line + attr.Reset, err
	}

	out     := "" // The final highlighted output
	lastCol := 0
	isCmd   := true
	for _, tok := range p.Toks {
		var next string
		var err  error
		lastCol, next, err = highlightNext(tok, line, lastCol, isCmd)
		if err != nil {
			return colorError + line + attr.Reset, err
		}

		out += next

		if isCmd {
			isCmd = false
		} else {
			if tok.Type == token.Separator {
				isCmd = true
			}
		}
	}

	// Color the comments
	out = strings.Replace(out, "#", colorComment + "#", -1)

	return out + attr.Reset, nil
}
