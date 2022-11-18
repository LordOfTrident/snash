package repl

import (
	"os"

	"github.com/LordOfTrident/snash/pkg/env"
	"github.com/LordOfTrident/snash/pkg/term"
	"github.com/LordOfTrident/snash/pkg/prompt"
	"github.com/LordOfTrident/snash/pkg/interpreter"
	"github.com/LordOfTrident/snash/pkg/highlighter"
)

func REPL(e *env.Env, interactive, showPossibleErrors bool) int {
	term.Init()

	p := prompt.New(interactive, showPossibleErrors)

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

	return e.Ex
}
