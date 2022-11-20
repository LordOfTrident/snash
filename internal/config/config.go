package config

import (
	"os"
	"flag"
	"fmt"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/highlighter"
)

const (
	AppName = "snash"

	VersionMajor = 1
	VersionMinor = 7
	VersionPatch = 5
)

var (
	Folder      = os.Getenv("HOME") + "/.config/snash"
	HistoryFile = Folder + "/history"
)

var (
	Interactive        = flag.Bool("interactive",        true, "Interactive REPL mode")
	ShowPossibleErrors = flag.Bool("showPossibleErrors", true, "Print the possible input errors")
	SyntaxHighlighting = flag.Bool("syntaxHighlighting", true, "Syntax highlight the input")
)

func HasFolder() bool {
	_, err := os.Stat(Folder);

	return err == nil
}

func FixFolder() bool {
	if !HasFolder() {
		highlighter.Printf("Config directory %v missing, creating it\n", utils.Quote(Folder))

		if err := os.Mkdir(Folder, os.ModePerm); err != nil {
			highlighter.PrintError(fmt.Errorf("Could not create config folder '%v/'", Folder))

			return false
		}
	}

	return true
}
