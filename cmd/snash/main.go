package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/LordOfTrident/snash/pkg/attr"
	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/repl"
	"github.com/LordOfTrident/snash/pkg/env"
	"github.com/LordOfTrident/snash/pkg/interpreter"
)

// 1.0.0: First release, executing simple commands
// 1.1.0: Added an interactive REPL
// 1.1.1: Add command line flags

const (
	appName = "snash"

	versionMajor = 1
	versionMinor = 1
	versionPatch = 1
)

var (
	interactive = flag.Bool("interactive", true, "Interactive REPL mode")
	showVersion = flag.Bool("version",     false, "Show the version")
)

func execScript(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		errors.Print(fmt.Errorf("Could not read file '%v'", path))

		os.Exit(1)
	}

	env := env.New()
	env.UpdateVars()

	err = interpreter.Interpret(env, string(data), path)
	if err != nil {
		errors.Print(err)

		os.Exit(1)
	}

	return env.Ex
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
	os.Setenv("ERR_PROMPT", "\\u@\\h \\w [\\[" + attr.Bold + attr.BrightRed + "\\]\\ex" +
	                        "\\[" + attr.Reset + "\\]] $ ")

	// Flag related things

	flag.Usage = usage

	// Aliases
	flag.BoolVar(showVersion, "v", *showVersion, "alias for -version")
	flag.BoolVar(interactive, "i", *interactive, "alias for -interactive")

	flag.Parse()
}

func main() {
	if *showVersion {
		version()

		return
	}

	if len(flag.Args()) > 1 {
		ex := 0

		// Range over the arguments that are not flags
		for _, arg := range flag.Args() {
			ex = execScript(arg)
		}

		os.Exit(ex)
	} else {
		os.Exit(repl.REPL(*interactive))
	}
}
