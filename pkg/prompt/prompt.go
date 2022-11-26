package prompt

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"unicode"
	"math"

	"github.com/LordOfTrident/snash/pkg/term"
)

// TODO: Multi-line prompt mode, which is gonna be like a text editor field inside the
//       command line

// TODO: Optimize the prompt rendering by making an output buffer system

type Highlighter interface {
	Highlight(code, path string) (string, error)
}

type History struct {
	list []string
	idx  int
}

func NewHistory() History {
	var h History
	h.list = append(h.list, "")

	return h
}

func LoadHistory(path string) (History, error) {
	f, err := os.Open(path)
	if err != nil {
		return NewHistory(), err
	}
	defer f.Close()

	h := NewHistory()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		h.Add(scanner.Text())
	}

	return h, nil
}

func (h *History) SaveToFile(path string) error {
	// Open history file for appending
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, v := range h.list {
		// Dont write the last empty line to the history, or any empty lines
		if len(v) == 0 {
			continue
		}

		f.WriteString(v + "\n")
	}

	return nil
}

func (h *History) Add(code string) {
	if len(code) > 0 {
		// Save input to history
		h.list[len(h.list) - 1] = code
		h.list = append(h.list, "")
	}

	return
}

func (h *History) Up() string {
	if h.idx > 0 {
		h.idx --
	}

	return h.list[h.idx]
}

func (h *History) Down() string {
	if h.idx < len(h.list) - 1 {
		h.idx ++
	}

	return h.list[h.idx]
}

func (h *History) ToEnd() {
	h.idx = len(h.list) - 1
}

func (p *Prompt) SetInput(input string) {
	*p.line = input
	p.curx  = len(*p.line)
}

type Prompt struct {
	History History

	Flags struct {
		Interactive, ShowPossibleErrors, SyntaxHighlighting bool
	}

	Colors struct {
		Error string
	}

	lines []string
	line   *string
	curx    int

	prevMode string // Previous terminal stty mode

	highlighter Highlighter
}

func New(h History, highlighter Highlighter) *Prompt {
	p := &Prompt{History: h, highlighter: highlighter}
	p.clear()

	// Default config
	p.Colors.Error = term.AttrGrey

	return p
}

func (p *Prompt) cursorChar() rune {
	if p.curx == len(*p.line) {
		return 0
	} else {
		return rune((*p.line)[p.curx])
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

func isWordChar(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_'
}

func (p *Prompt) moveCursorLeftByWord() {
	p.moveCursorLeft()

	for unicode.IsSpace(p.cursorChar()) {
		if p.curx == 0 {
			return
		}

		p.moveCursorLeft()
	}

	p.moveCursorLeft()
	for isWordChar(p.cursorChar()) {
		if p.curx == 0 {
			return
		}

		p.moveCursorLeft()
	}

	p.moveCursorRight()
}

func (p *Prompt) moveCursorRightByWord() {
	for unicode.IsSpace(p.cursorChar()) {
		p.moveCursorRight()

		if p.curx >= len(*p.line) {
			return
		}
	}

	for isWordChar(p.cursorChar()) {
		if p.curx >= len(*p.line) {
			p.moveCursorRight()

			return
		}

		p.moveCursorRight()
	}
}

func (p *Prompt) eraseCharAtCursor() {
	if p.curx > 0 {
		part1 := (*p.line)[:p.curx - 1]
		part2 := (*p.line)[p.curx:]

		*p.line = part1 + part2
		p.curx  --
	}
}

func (p *Prompt) insertCharAtCursor(char rune) {
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
		case 1:
			skip = true

			continue

		case 2:
			skip = false

			continue
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

func (p *Prompt) clear() {
	p.lines = []string{""}
	p.line  = &p.lines[0]
	p.curx  = 0

	p.History.ToEnd()
}

func (p *Prompt) Input(prompt string) string {
	var flags term.Flag
	flags = term.CBreak | term.NoEcho

	if p.Flags.Interactive {
		flags |= term.NoIxon
	}

	// Init the terminal
	p.prevMode = term.SaveMode()
	term.SetMode(flags)
	term.InitGetKey()
	term.Update()

	lastPromptLine, lastPromptLineLen := getLastPromptLine(prompt)

	// Remove the ignore marking characters
	prompt = strings.Replace(prompt, "\x01", "", -1)
	prompt = strings.Replace(prompt, "\x02", "", -1)

	fmt.Print(prompt) // Output all lines of the prompt (this is for multiline prompts)

	hasPossibleErrMsg    := false // Is there a possible error msg displayed?
	possibleErrLinesUsed := 1     // The number of lines a possible error was rendered over
	inputLinesUsed       := 1
	clearAll             := false

	// Hide the cursor for rendering
	term.HideCursor()

	typing := true
	for typing {
		var possibleErr      error
		var highlightedInput string

		// Highlight the input and show possible errors
		if p.Flags.Interactive && p.highlighter != nil {
			if p.Flags.ShowPossibleErrors || p.Flags.SyntaxHighlighting {
				highlightedInput, possibleErr = p.highlighter.Highlight(*p.line, "stdin")

				if !p.Flags.ShowPossibleErrors {
					possibleErr = nil
				}

				if !p.Flags.SyntaxHighlighting {
					highlightedInput = *p.line
				}
			}
		}

		// Clear the previous input
		if clearAll {
			term.ClearLines(inputLinesUsed)
		} else {
			term.NewLines(inputLinesUsed - 1)
			term.ClearCursorLine()
		}

		// Clear previous error msgs
		if hasPossibleErrMsg {
			term.NewLine()
			term.ClearLines(possibleErrLinesUsed)
			term.MoveCursorUp(possibleErrLinesUsed)
		}

		term.MoveCursorUp(inputLinesUsed - 1)

		// Output the last line of the prompt and the input
		fmt.Print(lastPromptLine, highlightedInput)

		// Create a new line for the cursor if only the cursor gets put on a new line
		if (lastPromptLineLen + len(*p.line)) % term.Width == 0 {
			term.NewLine()
			term.ClearCursorLine()
		}

		// Lines used by the prompt and input
		inputLinesUsed = (len(*p.line) + lastPromptLineLen) / term.Width + 1

		// Render possible errors if there are any
		if possibleErr != nil {
			possibleErrLinesUsed = p.renderPossibleError(possibleErr)

			hasPossibleErrMsg = true
		} else {
			hasPossibleErrMsg = false
		}

		// Position the cursor
		term.MoveCursorUp(inputLinesUsed - 1)
		term.MoveCursorToLineStart()

		offx := lastPromptLineLen + p.curx
		offy := 0
		if offx >= term.Width {
			offy = offx / term.Width
			offx = offx % term.Width

			term.NewLines(offy)
		}

		term.MoveCursorRight(offx)

		clearAll = false

		term.ShowCursor()

		// Read input
		key := term.GetKey(true)

		switch key {
		case term.KeyEnter: typing = false

		case term.KeyBackspace: p.eraseCharAtCursor()

		case term.KeyArrowUp:
			clearAll = true
			p.SetInput(p.History.Up())

		case term.KeyArrowDown:
			clearAll = true
			p.SetInput(p.History.Down())

		case term.KeyArrowRight: p.moveCursorRight()
		case term.KeyArrowLeft:  p.moveCursorLeft()

		case term.KeyCtrlArrowRight: p.moveCursorRightByWord()
		case term.KeyCtrlArrowLeft:  p.moveCursorLeftByWord()

		case term.Ctrl(term.Key('s')):

		case term.KeyResize:
			clearAll = true

			term.Update()

		default:
			if key >= term.Key(' ') && key <= term.Key('~') {
				p.insertCharAtCursor(rune(key))
			}
		}

		term.HideCursor()
		term.MoveCursorUp(offy)
	}

	term.NewLines(inputLinesUsed - 1)

	ret := ""
	for i, line := range p.lines {
		if i > 0 {
			ret += "; "
		}

		ret += line
	}

	p.History.Add(ret)

	fmt.Println()
	term.RestoreMode(p.prevMode)
	term.ShowCursor()

	p.clear()

	return ret
}

func (p *Prompt) renderPossibleError(err error) (linesUsed int) {
	term.NewLine()

	msg := "Error: " + err.Error()
	fmt.Print(p.Colors.Error + msg + term.AttrReset)

	linesUsed = int(math.Ceil(float64(len(msg)) / float64(term.Width)))

	term.MoveCursorUp(linesUsed)

	return
}
