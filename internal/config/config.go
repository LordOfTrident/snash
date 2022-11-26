package config

import (
	"os"
	"flag"
	"fmt"

	"github.com/LordOfTrident/snash/runtime"
)

const (
	AppName = "snash"

	VersionMajor = 1
	VersionMinor = 11
	VersionPatch = 7

	GithubLink = "https://github.com/LordOfTrident/snash"
)

var (
	Folder      = os.Getenv("HOME") + "/.config/snash/"
	HistoryPath = Folder + "history"
	RCPath      = Folder + runtime.RCFile
)

var (
	Interactive        = flag.Bool("interactive",        true, "Interactive REPL mode")
	ShowPossibleErrors = flag.Bool("showPossibleErrors", true, "Print the possible input errors")
	SyntaxHighlighting = flag.Bool("syntaxHighlighting", true, "Syntax highlight the input")
)

func CreateFolder() error {
	if err := os.Mkdir(Folder, os.ModePerm); err != nil {
		return fmt.Errorf("Could not create config folder")
	}

	return nil
}
