package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/LordOfTrident/snash/pkg/attr"
	"github.com/LordOfTrident/snash/pkg/repl"
	"github.com/LordOfTrident/snash/pkg/env"
	"github.com/LordOfTrident/snash/pkg/interpreter"
	"github.com/LordOfTrident/snash/pkg/highlighter"
)

// 1.0.0: First release, executing simple commands
// 1.1.0: Added an interactive REPL
// 1.1.1: Add command line flags
// 1.1.2: Ignore CTRL+C
// 1.2.2: Add an option to print possible input errors under the prompt
// 1.2.3: Improve syntax highlighting
// 1.2.4: Fix string escape sequences
// 1.2.5: Unescape error strings

const (
	appName = "snash"

	versionMajor = 1
	versionMinor = 2
	versionPatch = 5
)

var (
	showVersion = flag.Bool("version", false, "Show the version")

	interactive        = flag.Bool("interactive",        true, "Interactive REPL mode")
	showPossibleErrors = flag.Bool("showPossibleErrors", true, "Print the possible input errors")
)

var e = env.New()

func execScript(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		highlighter.PrintError(fmt.Errorf("Could not read file '%v'", path))

		os.Exit(1)
	}

	e.UpdateVars()

	err = interpreter.Interpret(e, string(data), path)
	if err != nil {
		highlighter.PrintError(err)

		os.Exit(1)
	}

	return e.Ex
}

func usage() {
	fmt.Printf("Usage: %v [FILE...] [OPTIONS]\n", os.Args[0])
	fmt.Println("Options:")

	flag.PrintDefaults()
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", appName, versionMajor, versionMinor, versionPatch)
}

func init() {
	// Defaults
	os.Setenv("SHELL", os.Args[0])

	os.Setenv("PROMPT",     "\\u@\\h \\w $ ")
	os.Setenv("PROMPT_ERR", "\\u@\\h \\w [\\[" + attr.Bold + attr.BrightRed + "\\]\\ex" +
	                        "\\[" + attr.Reset + "\\]] $ ")
	os.Setenv("PROMPT_MULTILINE", "> ")

	// Flag related things

	flag.Usage = usage

	// Aliases
	flag.BoolVar(showVersion, "v", *showVersion, "alias for -version")

	flag.Parse()
}

func main() {
	if *showVersion {
		version()

		return
	}

	if len(flag.Args()) > 0 {
		ex := 0

		// Range over the arguments that are not flags
		for _, arg := range flag.Args() {
			ex = execScript(arg)
		}

		os.Exit(ex)
	} else {
		os.Exit(repl.REPL(e, *interactive, *showPossibleErrors))
	}
}
