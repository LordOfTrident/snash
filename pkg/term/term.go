package term

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
)

var mode string

var (
	Width  = 0
	Height = 0
)

type Key int
const (
	// Additional keys
	KeyArrowUp    = Key(iota + 256)
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight

	KeyCtrlArrowUp
	KeyCtrlArrowDown
	KeyCtrlArrowLeft
	KeyCtrlArrowRight

	KeyResize // Terminal window resized event

	// ASCII keys
	KeyEnter     = Key('\n')
	KeyBackspace = Key(127)
	KeyEscape    = Key(27)
	KeyTab       = Key('\t')
)

func Init() {
	// Ignore CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
	}()

	// Save the previous terminal attributes
	bytes, err := exec.Command("stty", "-F", "/dev/tty", "-g").Output()
	if err != nil {
		panic(err)
	}

	mode = string(bytes[:len(bytes) - 1])
}

func RestoreMode() {
	// Restore the previous terminal attributes
	exec.Command("stty", "-F", "/dev/tty", mode).Run()
}

func InputMode() {
	// Some terminal attributes so we can easily read each key press
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
}

// 'stty size' output format is '<HEIGHT> <WIDTH>'
var sizeRegex = regexp.MustCompile("([0-9]*)\\s([0-9]*)\\s*")

func Update() {
	bytes, err := exec.Command("stty", "-F", "/dev/tty", "size").Output()
	if err != nil {
		panic(err)
	}

	// Match the output format
	info := sizeRegex.FindStringSubmatch(string(bytes))

	// Parse the strings
	Height, err = strconv.Atoi(info[1])
	if err != nil {
		panic(err)
	}

	Width, err = strconv.Atoi(info[2])
	if err != nil {
		panic(err)
	}
}

func MoveCursorToLineStart() {
	fmt.Print("\r")
}

func ClearCursorLine() {
	if Width <= 0 {
		Update()
	}

	fmt.Printf("\r%v\r", strings.Repeat(" ", Width - 1))
}

func MoveCursorUp(by int) {
	fmt.Printf("\x1b[%vA", by)
}

func MoveCursorDown(by int) {
	fmt.Printf("\x1b[%vB", by)
}

func MoveCursorLeft(by int) {
	fmt.Printf("\x1b[%vC", by)
}

func MoveCursorRight(by int) {
	fmt.Printf("\x1b[%vD", by)
}

func NewLine() {
	fmt.Print("\n")
}

func GetKey() (key Key) {
	prevWidth, prevHeight := Width, Height
	Update()

	// If the width/height changed, the window was resized
	if prevWidth != Width || prevHeight != Height {
		return KeyResize
	}

	// Array big enough to catch escape sequences that we need
	in := make([]byte, 8)
	os.Stdin.Read(in)

	switch in[0] {
	case 27:
		// Find the length of the sequence
		length := 0
		for _, v := range in {
			if v == 0 {
				break
			}

			fmt.Printf("%v ", v)

			length ++
		}

		switch length {
		case 1: key = KeyEscape // Just the escape key
		case 3: // Arrow keys sequence
			switch in[2] {
			case 'A': key = KeyArrowUp
			case 'B': key = KeyArrowDown
			case 'C': key = KeyArrowRight
			case 'D': key = KeyArrowLeft
			}

		case 6: // CTRL + arrow keys sequence
			if in[2] == 49 && in[3] == 59 && in[4] == 53 {
				switch in[5] {
				case 'A': key = KeyCtrlArrowUp
				case 'B': key = KeyCtrlArrowDown
				case 'C': key = KeyCtrlArrowRight
				case 'D': key = KeyCtrlArrowLeft
				}
			}
		}

	default: key = Key(in[0]) // Otherwise, return the first character
	}

	return
}
