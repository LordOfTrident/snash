package prompt

import (
	"fmt"

	"github.com/LordOfTrident/snash/pkg/attr"
	"github.com/LordOfTrident/snash/pkg/term"
	"github.com/LordOfTrident/snash/pkg/highlighter"
)

type Prompt struct {
	// Flags
	Interactive, ShowPossibleErrors bool

	history    []string
	historyIdx int

	input string
	curx  int

	prompt, promptLastLine string
	promptLastLineLen      int

	prevErrored bool // Did the previous input have an error?
}

func New(interactive, showPossibleErrors bool) *Prompt {
	p := &Prompt{Interactive: interactive, ShowPossibleErrors: showPossibleErrors, historyIdx: 0}
	p.history = append(p.history, "")

	return p
}

func (p *Prompt) historyUp() {
	if p.historyIdx > 0 {
		p.historyIdx --
	}

	// Reset the cursor position and update the input if we moved in history
	p.input = p.history[p.historyIdx]
	p.curx  = len(p.input)
}

func (p *Prompt) historyDown() {
	if p.historyIdx < len(p.history) - 1 {
		p.historyIdx ++
	}

	// Reset the cursor position and update the input if we moved in history
	p.input = p.history[p.historyIdx]
	p.curx  = len(p.input)
}

func (p *Prompt) finishInput() {
	if len(p.input) > 0 {
		// Save input to history
		p.history[len(p.history) - 1] = p.input
		p.history = append(p.history, "")

		p.input = ""
	}
}

func (p *Prompt) moveCursorLeft() {
	if p.curx > 0 {
		p.curx --
	}
}

func (p *Prompt) moveCursorRight() {
	if p.curx < len(p.input) {
		p.curx ++
	}
}

func (p *Prompt) eraseCursorChar() {
	if p.curx > 0 {
		part1 := p.input[:p.curx - 1]
		part2 := p.input[p.curx:]

		p.input = part1 + part2
		p.curx  --
	}
}

func (p *Prompt) insertCharAtCursor(char byte) {
	// Save the input left and right parts to insert the character
	part1 := p.input[:p.curx]
	part2 := p.input[p.curx:]

	p.input = part1 + string(char) + part2
	p.curx  ++
}

func (p *Prompt) setPrompt(prompt string) {
	p.prompt = prompt

	p.promptLastLineLen = 0
	p.promptLastLine    = ""

	// Calculate the length of the last prompt line
	skip := false
	for _, ch := range prompt {
		// Ignore characters marked to be ignored
		switch ch {
		case 1: skip = true
		case 2: skip = false
		}

		if !skip {
			if ch == '\n' {
				p.promptLastLineLen = 0
				p.promptLastLine    = ""
			} else {
				p.promptLastLineLen ++
				p.promptLastLine    += string(ch)
			}
		}
	}
}

func (p *Prompt) renderPossibleErrorLine(err error) {
	if err != nil {
		term.NewLine()
		term.ClearCursorLine()
		fmt.Printf("%vError: %v%v", attr.Grey, err.Error(), attr.Reset)
		term.MoveCursorUp(1)

		p.prevErrored = true
	} else if p.prevErrored {
		term.MoveCursorDown(1)
		term.ClearCursorLine()
		term.MoveCursorUp(1)

		p.prevErrored = false
	}
}

func (p *Prompt) clear() {
	p.input       = ""
	p.curx        = 0
	p.historyIdx  = len(p.history) - 1
	p.prevErrored = false
}

func (p *Prompt) Input(prompt string) string {
	term.InputMode()

	p.clear()

	p.setPrompt(prompt)
	fmt.Print(prompt) // Output all lines of the prompt

loop:
	for {
		var out string
		if p.Interactive {
			var err error
			out, err = highlighter.HighlightLine(p.input, "stdin")
			if p.ShowPossibleErrors {
				p.renderPossibleErrorLine(err)
			}
		} else {
			out = p.input
		}

		term.ClearCursorLine()

		// Output the last line of the prompt and the input
		fmt.Print(p.promptLastLine, out)

		// Position the cursor
		offx := len(p.input) - p.curx
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

		case term.KeyResize: fmt.Println("RESIZED")

		default:
			if key >= term.Key(' ') && key <= term.Key('~') {
				p.insertCharAtCursor(byte(key))
			}
		}
	}

	// Save the input to return it
	ret := p.input

	p.finishInput()
	fmt.Println()

	term.RestoreMode()

	return ret
}
