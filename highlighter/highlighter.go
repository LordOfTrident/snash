package highlighter

import (
	"os"
	"os/exec"
	"strings"

	"github.com/LordOfTrident/snash/attr"
	"github.com/LordOfTrident/snash/token"
	"github.com/LordOfTrident/snash/parser"
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

// TODO: Use the nodes to highlight - modify all nodes and args to contain the tokens so its
//       possible to use them for highlighting

func highlightNext(tok token.Token, line string, lastCol int, isCmd bool) (int, string) {
	idx := tok.Where.Col - 1

	var spaces, highlighted string
	if idx - lastCol > 0 {
		spaces = line[lastCol:idx]
	}

	if tok.Type == token.EOF {
		highlighted += spaces
	} else {
		txt := line[idx:idx + tok.TxtLen]

		lastCol = idx + tok.TxtLen

		switch tok.Type {
		case token.Integer: highlighted += spaces + colorInteger + txt

		default:
			if tok.Type.IsKeyword() {
				highlighted += spaces + colorKeyword + txt
			} else if isCmd && cmdExists(tok.Data) {
				highlighted += spaces + colorCmd + txt
			} else if !isCmd && fileExists(tok.Data){
				highlighted += spaces + colorPath + txt
			} else {
				highlighted += spaces + txt
			}
		}

		highlighted += attr.Reset
	}

	return lastCol, highlighted
}

func HighlightLine(line, path string) string {
	p, err := parser.New(line, path)
	if err != nil {
		return colorError + line + attr.Reset
	}

	_, err = p.Parse()
	if err != nil {
		return colorError + line + attr.Reset
	}

	out     := ""
	lastCol := 0
	isCmd   := true
	for _, tok := range p.Toks {
		var next string
		lastCol, next = highlightNext(tok, line, lastCol, isCmd)

		out += next

		if isCmd {
			isCmd = false
		} else {
			if tok.Type == token.Separator {
				isCmd = true
			}
		}
	}

	out = strings.Replace(out, "#", colorComment + "#", -1)

	return out + attr.Reset
}
