package main

import (
	"os"
	"fmt"

	"github.com/LordOfTrident/snash/attr"
	"github.com/LordOfTrident/snash/prompt"
	"github.com/LordOfTrident/snash/env"
	"github.com/LordOfTrident/snash/interpreter"
)

// 1.0.0: First release, executing simple commands
// 1.1.0: Added an interactive REPL

// App info
const (
	appName = "snash"

	versionMajor = 1
	versionMinor = 0
	versionPatch = 0
)

func printError(err error) {
	fmt.Fprintf(os.Stderr, "%v%vError:%v %v\n", attr.Bold, attr.BrightRed, attr.Reset, err.Error())
}

func repl() int {
	prompt.Init()

	env := env.New()
	p   := prompt.New()

	for {
		env.UpdateVars()

		// Generate a prompt
		var prompt string
		if env.Ex == 0 {
			prompt = env.GenPrompt(os.Getenv("PROMPT"))
		} else {
			prompt = env.GenPrompt(os.Getenv("ERR_PROMPT"))
		}

		// TODO: make an interactive flag variable
		in := p.ReadLine(prompt, true)

		err := interpreter.Interpret(env, in, "stdin")
		if err != nil {
			printError(err)
		}

		// Exit the repl if last exit was forced
		if env.Flags.ForcedExit {
			break
		}
	}

	return env.Ex
}

func fromFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		printError(fmt.Errorf("Could not read file '%v'", path))

		os.Exit(1)
	}

	env := env.New()
	env.UpdateVars()

	err = interpreter.Interpret(env, string(data), path)
	if err != nil {
		printError(err)

		os.Exit(1)
	}

	return env.Ex
}

func usage() {
	fmt.Printf("Usage: %v [FILE...]\n", os.Args[0])
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", appName, versionMajor, versionMinor, versionPatch)
}

func init() {
	// Defaults
	os.Setenv("SHELL", os.Args[0])

	os.Setenv("PROMPT",     "\\u@\\h \\w $ ")
	os.Setenv("ERR_PROMPT", "\\u@\\h \\w [\\[" + attr.Bold + attr.BrightRed + "\\]\\ex" +
	                        "\\[" + attr.Reset + "\\]] $ ")
}

func main() {
	if len(os.Args) > 1 {
		ex := 0

		// Range over the arguments, skipping the first one which is the program path
		for _, arg := range os.Args[1:] {
			switch arg {
			case "-h", "--help":    usage();   return
			case "-v", "--version": version(); return

			default: ex = fromFile(arg)
			}
		}

		os.Exit(ex)
	} else {
		os.Exit(repl())
	}
}
