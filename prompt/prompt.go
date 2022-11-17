package prompt

import (
	"os"
	"os/exec"
	"fmt"
	"strings"

	"github.com/LordOfTrident/snash/highlighter"
)

var mode string

type Prompt struct {
	history    []string
	historyIdx int
}

func Init() {
	// Save the previous terminal attributes
	bytes, err := exec.Command("stty", "-F", "/dev/tty", "-g").Output()
	if err != nil {
		panic(err)
	}

	mode = string(bytes[:len(bytes) - 1])
}

func begin() {
	// Some terminal attributes so we can easily read each key press
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
}

func end() {
	// Restore the previous terminal attributes
	exec.Command("stty", "-F", "/dev/tty", mode).Run()
}

func New() *Prompt {
	p := &Prompt{historyIdx: 0}
	p.history = append(p.history, "")

	return p
}

func (p *Prompt) historyEnd() int {
	return len(p.history) - 1
}

func (p *Prompt) historyUp() {
	if p.historyIdx > 0 {
		p.historyIdx --
	}
}

func (p *Prompt) historyDown() {
	if p.historyIdx < p.historyEnd() {
		p.historyIdx ++
	}
}

func (p *Prompt) historyAdd(line string) {
	p.history = append(p.history, line)
}

func (p *Prompt) ReadLine(prompt string, interactive bool) (input string) {
	begin()

	skip      := false
	promptLen := 0 // The length of the last line of the prompt
	newLines       := 0  // The amount of new lines in the prompt
	promptLastPart := "" // The last line of the prompt

	// Calculate promptLen
	for _, v := range prompt {
		// Ignore  characters marked to be ignored
		switch v {
		case 1: skip = true
		case 2: skip = false
		}

		if !skip {
			if v == '\n' {
				newLines ++

				promptLen      = 0
				promptLastPart = ""
			} else {
				promptLen      ++
				promptLastPart += string(v)
			}
		}
	}

	// Array big enough to catch escape sequences that we need
	in   := make([]byte, 8)
	curx := 0

	p.historyIdx = p.historyEnd()

	// The previous length of input (for clearing the line with whitespaces)
	prevLen := 0

	fmt.Print(prompt) // Output all lines of the prompt

loop:
	for {
		var out string
		if interactive {
			out = highlighter.HighlightLine(input, "stdin")
		} else {
			out = input
		}

		// Clear the line
		fmt.Printf("\r%v\r", strings.Repeat(" ", promptLen + prevLen))

		// Output the last line of the prompt and the input
		fmt.Print(promptLastPart, out)

		// Position the cursor
		offx := len(input) - curx
		if offx != 0 {
			fmt.Printf("\x1b[%vD", offx)
		}

		prevLen = len(input)

		// Read input
		os.Stdin.Read(in)
		ch := in[0]

		switch ch {
		case '\n': break loop

		case 127:
			if curx > 0 {
				part1 := input[:curx - 1]
				part2 := input[curx:]

				input = part1 + part2
				curx  --
			}

		case 27:
			// Find the length of the sequence
			length := 0
			for _, v := range in {
				if v == 0 {
					break
				}

				length ++
			}

			// Read the arrow keys sequence
			if length == 3 {
				switch in[2] {
				case 'A': p.historyUp()
				case 'B': p.historyDown()

				case 'C': // Right
					if curx < len(input) {
						curx ++
					}

				case 'D': // Left
					if curx > 0 {
						curx --
					}
				}

				// Reset the cursor position and update the input if we moved in history
				if in[2] == 'A' || in[2] == 'B' {
					input = p.history[p.historyIdx]
					curx  = len(input)
				}
			}

		default:
			if ch >= ' ' && ch <= '~' {
				// Save the input left and right parts to insert the character
				part1 := input[:curx]
				part2 := input[curx:]

				input = part1 + string(ch) + part2
				curx  ++
			}
		}
	}

	// Save input to history
	if len(input) > 0 {
		p.history[p.historyEnd()] = input
		p.historyAdd("")
	}

	fmt.Println()

	end()

	return
}
