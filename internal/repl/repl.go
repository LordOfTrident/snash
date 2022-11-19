package repl

import (
	"os"

	"github.com/LordOfTrident/snash/internal/config"
	"github.com/LordOfTrident/snash/internal/env"
	"github.com/LordOfTrident/snash/internal/term"
	"github.com/LordOfTrident/snash/internal/prompt"
	"github.com/LordOfTrident/snash/internal/interpreter"
	"github.com/LordOfTrident/snash/internal/highlighter"
)

func REPL(e *env.Env) int {
	term.Init()

	p := prompt.New(*config.Interactive, *config.ShowPossibleErrors)

	for {
		e.UpdateVars()

		// Generate a prompt
		var prompt string
		if e.Ex == 0 {
			prompt = e.GenPrompt(os.Getenv("PROMPT"))
		} else {
			prompt = e.GenPrompt(os.Getenv("PROMPT_ERR"))
		}

		in := p.Input(prompt, e.GenPrompt(os.Getenv("PROMPT_MULTILINE")))

		err := interpreter.Interpret(e, in, "stdin")
		if err != nil {
			highlighter.PrintError(err)
		}

		// Exit the repl if last exit was forced
		if e.Flags.ForcedExit {
			break
		}
	}

	p.SaveHistoryToFile()

	return e.Ex
}
