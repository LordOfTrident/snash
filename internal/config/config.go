package config

import (
	"os"
	"flag"
	"fmt"
)

const (
	AppName = "snash"

	VersionMajor = 1
	VersionMinor = 10
	VersionPatch = 6

	GithubLink = "https://github.com/LordOfTrident/snash"
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

func FolderExists() bool {
	_, err := os.Stat(Folder);

	return err == nil
}

func CreateFolder() error {
	if err := os.Mkdir(Folder, os.ModePerm); err != nil {
		return fmt.Errorf("Could not create config folder")
	}

	return nil
}
