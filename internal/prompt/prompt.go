package prompt

import (
	"fmt"
	"os"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/config"
	"github.com/LordOfTrident/snash/internal/attr"
	"github.com/LordOfTrident/snash/internal/term"
	"github.com/LordOfTrident/snash/internal/highlighter"
)

// TODO: Multi-line prompt mode, which is gonna be like a text editor field inside the
//       command line. Pressing ESC will execute that code, and pressing CTRL+S will save it into
//       a chosen file.

type Prompt struct {
	// Flags
	Interactive, ShowPossibleErrors bool

	history    []string
	historyIdx int

	lines []string
	line   *string
	curx    int

	prevErrored bool // Did the previous input have an error?
}

func New(interactive, showPossibleErrors bool) *Prompt {
	p := &Prompt{Interactive: interactive, ShowPossibleErrors: showPossibleErrors, historyIdx: 0}
	p.history = append(p.history, "")

	return p
}

func (p *Prompt) SaveHistoryToFile() {
	if !config.HasFolder() {
		return
	}

	// Open history file for appending
	f, err := os.OpenFile(config.HistoryFile, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
	if err != nil {
		highlighter.PrintError(fmt.Errorf("Could not save history file '%v'", config.HistoryFile))
	}
	defer f.Close()

	for _, v := range p.history {
		// Dont write the last empty line to the history, or any empty lines
		if len(v) == 0 {
			continue
		}

		f.WriteString(v + "\n")
	}
}

func (p *Prompt) historyUp() {
	if p.historyIdx > 0 {
		p.historyIdx --
	}

	// Reset the cursor position and update the input if we moved in history
	*p.line = p.history[p.historyIdx]
	p.curx  = len(*p.line)
}

func (p *Prompt) historyDown() {
	if p.historyIdx < len(p.history) - 1 {
		p.historyIdx ++
	}

	// Reset the cursor position and update the input if we moved in history
	*p.line = p.history[p.historyIdx]
	p.curx  = len(*p.line)
}

func (p *Prompt) finishInput() (input string) {
	for i, line := range p.lines {
		if i > 0 {
			input += "; "
		}

		input += line
	}

	if len(input) > 0 {
		// Save input to history
		p.history[len(p.history) - 1] = input
		p.history = append(p.history, "")
	}

	return
}

func (p *Prompt) cursorChar() byte {
	if p.curx == len(*p.line) {
		return utils.CharNone
	} else {
		return (*p.line)[p.curx]
	}
}

func (p *Prompt) moveCursorLeft() {
	if p.curx > 0 {
		p.curx --
	}
}

func (p *Prompt) moveCursorRight() {
	if p.curx < len(*p.line) {
		p.curx ++
	}
}

func (p *Prompt) moveCursorLeftByWord() {
	p.moveCursorLeft()

	for utils.IsWhitespace(p.cursorChar()) {
		if p.curx == 0 {
			return
		}

		p.moveCursorLeft()
	}

	p.moveCursorLeft()
	for utils.IsAlphanum(p.cursorChar()) || p.cursorChar() == '_' {
		if p.curx == 0 {
			return
		}

		p.moveCursorLeft()
	}

	p.moveCursorRight()
}

func (p *Prompt) moveCursorRightByWord() {
	for utils.IsWhitespace(p.cursorChar()) {
		p.moveCursorRight()

		if p.curx >= len(*p.line) {
			return
		}
	}

	for utils.IsAlphanum(p.cursorChar()) || p.cursorChar() == '_' {
		if p.curx >= len(*p.line) {
			p.moveCursorRight()

			return
		}

		p.moveCursorRight()
	}
}

func (p *Prompt) eraseCursorChar() {
	if p.curx > 0 {
		part1 := (*p.line)[:p.curx - 1]
		part2 := (*p.line)[p.curx:]

		*p.line = part1 + part2
		p.curx  --
	}
}

func (p *Prompt) insertCharAtCursor(char byte) {
	// Save the input left and right parts to insert the character
	part1 := (*p.line)[:p.curx]
	part2 := (*p.line)[p.curx:]

	*p.line = part1 + string(char) + part2
	p.curx  ++
}

func getLastPromptLine(prompt string) (lastLine string, lastLineLen int) {
	skip := false
	for _, ch := range prompt {
		// Ignore characters marked to be ignored
		switch ch {
		case 1: skip = true
		case 2: skip = false
		}

		lastLine += string(ch)

		if !skip {
			if ch == '\n' {
				lastLineLen = 0
				lastLine    = ""
			} else {
				lastLineLen ++
			}
		}
	}

	return
}

func (p *Prompt) renderPossibleErrorLine(err error) {
	if err != nil {
		// Render the error line
		term.NewLine()
		term.ClearCursorLine()
		fmt.Printf("%vError: %v%v", attr.Grey, err.Error(), attr.Reset)
		term.MoveCursorUp(1)

		p.prevErrored = true
	} else if p.prevErrored {
		// Clear the error line
		term.MoveCursorDown(1)
		term.ClearCursorLine()
		term.MoveCursorUp(1)

		p.prevErrored = false
	}
}

func (p *Prompt) clear() {
	p.lines       = []string{""}
	p.line        = &p.lines[0]
	p.curx        = 0
	p.historyIdx  = len(p.history) - 1
	p.prevErrored = false
}

func (p *Prompt) Input(prompt, multiLinePrompt string) string {
	term.InputMode()

	p.clear()

	promptLastLine, _ := getLastPromptLine(prompt)
	//promptLastLine, promptLastLineLen := getLastPromptLine(prompt)

	fmt.Print(prompt) // Output all lines of the prompt

loop:
	for {
		var out string
		if p.Interactive {
			var err error
			out, err = highlighter.HighlightLine(*p.line, "stdin")
			if p.ShowPossibleErrors {
				p.renderPossibleErrorLine(err)
			}
		} else {
			out = *p.line
		}

		term.ClearCursorLine()

		// Output the last line of the prompt and the input
		fmt.Print(promptLastLine, out)

		// Position the cursor
		offx := len(*p.line) - p.curx
		if offx != 0 {
			term.MoveCursorRight(offx)
		}

		// Read input
		key := term.GetKey()

		switch key {
		case term.KeyEnter: break loop

		case term.KeyBackspace: p.eraseCursorChar()

		case term.KeyArrowUp:    p.historyUp()
		case term.KeyArrowDown:  p.historyDown()
		case term.KeyArrowRight: p.moveCursorRight()
		case term.KeyArrowLeft:  p.moveCursorLeft()

		case term.KeyCtrlArrowRight: p.moveCursorRightByWord()
		case term.KeyCtrlArrowLeft:  p.moveCursorLeftByWord()

		case term.KeyResize:

		case term.Ctrl(term.Key('s')):

		default:
			if key >= term.Key(' ') && key <= term.Key('~') {
				p.insertCharAtCursor(byte(key))
			}
		}
	}

	ret := p.finishInput()
	fmt.Println()

	term.RestoreMode()

	return ret
}
