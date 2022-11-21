package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/LordOfTrident/snash/internal/config"
	"github.com/LordOfTrident/snash/internal/repl"
	"github.com/LordOfTrident/snash/internal/env"
	"github.com/LordOfTrident/snash/internal/evaluator"
	"github.com/LordOfTrident/snash/internal/highlighter"
	"github.com/LordOfTrident/snash/internal/term"
)

// 1.0.0: First release, executing simple commands
// 1.1.0: Added an interactive REPL
// 1.2.0: Add command line flags
// 1.2.1: Ignore CTRL+C
// 1.3.1: Add an option to print possible input errors under the prompt
// 1.3.2: Improve syntax highlighting
// 1.3.3: Fix string escape sequences
// 1.3.4: Unescape error strings
// 1.4.4: CTRL + arrow keys cursor movement
// 1.5.4: Config folder + REPL history file
// 1.6.4: Syntax highlighting flag, make ixon not be disabled when the mode is not interactive
// 1.6.5: Remove ignore marker characters from the prompt when it is rendered
// 1.7.5: Add logical and, or operators
// 1.7.6: Fix logical or (it was lexed as a logical and)

var showVersion = flag.Bool("version", false, "Show the version")

var e = env.New()

func execScript(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		highlighter.PrintError(fmt.Errorf("Could not read file '%v'", path))

		os.Exit(1)
	}

	e.UpdateVars()

	err = evaluator.Eval(e, string(data), path)
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
	fmt.Printf("%v %v.%v.%v\n",
	           config.AppName, config.VersionMajor, config.VersionMinor, config.VersionPatch)
}

func init() {
	// Defaults
	os.Setenv("SHELL", os.Args[0])

	os.Setenv("PROMPT",     "\\u@\\h \\w $ ")
	os.Setenv("PROMPT_ERR", "\\u@\\h \\w [\\[" + term.AttrBold + term.AttrBrightRed + "\\]\\ex" +
	                        "\\[" + term.AttrReset + "\\]] $ ")
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

	config.FixFolder()

	if len(flag.Args()) > 0 {
		ex := 0

		// Range over the arguments that are not flags
		for _, arg := range flag.Args() {
			ex = execScript(arg)
		}

		os.Exit(ex)
	} else {
		os.Exit(repl.REPL(e))
	}
}
