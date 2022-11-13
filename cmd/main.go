package main

import (
	"os"
	"fmt"
	"bufio"

	"github.com/LordOfTrident/snash/interpreter"
)

const (
	appName = "gastroshell"

	versionMajor = 1
	versionMinor = 0
	versionPatch = 0
)

var stdin *bufio.Reader

func repl() int {
	ex := 0

	for {
		fmt.Print("$ ")

		in, err := stdin.ReadString('\n')
		if err != nil {
			panic(err)
		}

		var forced bool
		ex, forced, err = interpreter.Interpret(in, "stdin")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
		}

		if forced {
			break
		}
	}

	return ex
}

func fromFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 1, fmt.Errorf("Could not read file '%v'", path)
	}

	ex, _, err := interpreter.Interpret(string(data), path)
	if err != nil {
		return 1, err
	}

	return ex, nil
}

func usage() {
	fmt.Printf("Usage: %v [FILE...]\n", os.Args[0])
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", appName, versionMajor, versionMinor, versionPatch)
}

func init() {
	os.Setenv("SHELL", os.Args[0])

	stdin = bufio.NewReader(os.Stdin)
}

func main() {
	lastEx := 0

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			switch arg {
			case "-h", "--help":    usage();   return
			case "-v", "--version": version(); return

			default:
				ex, err := fromFile(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())

					os.Exit(1)
				}

				lastEx = ex
			}
		}
	} else {
		lastEx = repl()
	}

	os.Exit(lastEx)
}
