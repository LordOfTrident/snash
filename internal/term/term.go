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

const (
	AttrReset     = "\x1b[0m"
	AttrBold      = "\x1b[1m"
	AttrItalics   = "\x1b[3m"
	AttrUnderline = "\x1b[4m"
	AttrBlink     = "\x1b[5m"

	AttrBlack   = "\x1b[30m"
	AttrRed     = "\x1b[31m"
	AttrGreen   = "\x1b[32m"
	AttrYellow  = "\x1b[33m"
	AttrBlue    = "\x1b[34m"
	AttrMagenta = "\x1b[35m"
	AttrCyan    = "\x1b[36m"
	AttrWhite   = "\x1b[37m"

	AttrGrey          = "\x1b[90m"
	AttrBrightRed     = "\x1b[91m"
	AttrBrightGreen   = "\x1b[92m"
	AttrBrightYellow  = "\x1b[93m"
	AttrBrightBlue    = "\x1b[94m"
	AttrBrightMagenta = "\x1b[95m"
	AttrBrightCyan    = "\x1b[96m"
	AttrBrightWhite   = "\x1b[97m"
)

var (
	Width  = 0
	Height = 0
)

var mode string

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

func Ctrl(key Key) Key {
	switch key {
	case KeyArrowUp:    return KeyCtrlArrowUp
	case KeyArrowDown:  return KeyCtrlArrowDown
	case KeyArrowLeft:  return KeyCtrlArrowLeft
	case KeyArrowRight: return KeyCtrlArrowRight

	default: return Key(int(key) & 31)
	}
}

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
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run() // No input echo
	exec.Command("stty", "-F", "/dev/tty", "-ixon").Run() // Recieve CTRL+S
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
				case 'A': key = Ctrl(KeyArrowUp)
				case 'B': key = Ctrl(KeyArrowDown)
				case 'C': key = Ctrl(KeyArrowRight)
				case 'D': key = Ctrl(KeyArrowLeft)
				}
			}
		}

	default: key = Key(in[0]) // Otherwise, return the first character
	}

	return
}
