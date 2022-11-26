package term

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"regexp"
	"strconv"
	"strings"
	"time"
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

type Key int
const (
	// Additional keys
	KeyArrowUp = Key(iota + 256)
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight

	KeyCtrlArrowUp
	KeyCtrlArrowDown
	KeyCtrlArrowLeft
	KeyCtrlArrowRight

	KeyResize // Window resize event

	// ASCII keys
	KeyEnter     = Key('\n')
	KeyBackspace = Key(127)
	KeyEscape    = Key(27)
	KeyTab       = Key('\t')

	KeyNone = 0
)

type Flag int
const (
	CBreak = 1 << iota
	NoEcho
	Echo
	NoIxon
	Ixon
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

func SaveMode() string {
	// Save the previous terminal attributes
	bytes, err := exec.Command("stty", "-F", "/dev/tty", "-g").Output()
	if err != nil {
		panic(err)
	}

	return string(bytes[:len(bytes) - 1])
}

func RestoreMode(mode string) {
	// Restore the previous terminal attributes
	exec.Command("stty", "-F", "/dev/tty", mode).Run()
}

func OnCtrlC(callback func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)

	go func() {
		for {
			<- c
			callback()
		}
	}()
}

var resizedEvent = false

func SendResizeEvents() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)

	go func() {
		for {
			<- c
			resizedEvent = true
		}
	}()
}

func InitGetKey() {
	exec.Command("stty", "-F", "/dev/tty", "min", "0", "time", "0").Run()
}

func SetMode(flags Flag) {
	args := []string{"-F", "/dev/tty"}

	if flags & CBreak != 0 {
		args = append(args, "cbreak")
	}

	if flags & NoEcho != 0 {
		args = append(args, "-echo")
	} else if flags & Echo != 0 {
		args = append(args, "+echo")
	}

	// Recieve CTRL+S etc
	if flags & NoIxon != 0 {
		args = append(args, "-ixon")
	} else if flags & Ixon != 0 {
		args = append(args, "+ixon")
	}

	exec.Command("stty", args...).Run()
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

func HideCursor() {
	fmt.Print("\x1b[?25l")
}

func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

func MoveCursorToLineStart() {
	fmt.Print("\r")
}

func ClearCursorLine() {
	if Width <= 0 {
		Update()
	}

	fmt.Printf("\r%v\r", strings.Repeat(" ", Width))
}

func MoveCursorUp(by int) {
	if by == 0 {
		return
	} else if by < 0 {
		fmt.Printf("\x1b[%vB", -by)
	} else {
		fmt.Printf("\x1b[%vA", by)
	}
}

func MoveCursorDown(by int) {
	if by == 0 {
		return
	} else if by < 0 {
		fmt.Printf("\x1b[%vA", -by)
	} else {
		fmt.Printf("\x1b[%vB", by)
	}
}

func MoveCursorRight(by int) {
	if by == 0 {
		return
	} else if by < 0 {
		fmt.Printf("\x1b[%vD", -by)
	} else {
		fmt.Printf("\x1b[%vC", by)
	}
}

func MoveCursorLeft(by int) {
	if by == 0 {
		return
	} else if by < 0 {
		fmt.Printf("\x1b[%vC", -by)
	} else {
		fmt.Printf("\x1b[%vD", by)
	}
}

func NewLine() {
	fmt.Print("\n")
}

func NewLines(count int) {
	for i := 0; i < count; i ++ {
		NewLine()
	}
}

func ClearLines(count int) {
	for i := 0; i < count; i ++ {
		if i > 0 {
			NewLine()
		}

		ClearCursorLine()
	}
}

func GetKey(blocking bool) (key Key) {
	// Array big enough to catch escape sequences that we need
	in := make([]byte, 8)

	// I hope this infinite loop waiting for a keypress/event wont cause any issues
	for {
		os.Stdin.Read(in)

		if in[0] == 0 {
			if resizedEvent {
				resizedEvent = false

				return KeyResize
			} else if !blocking {
				return KeyNone
			}
		} else {
			break
		}

		time.Sleep(20 * time.Millisecond)
	}

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
