package repl

import (
	"os"

	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/env"
	"github.com/LordOfTrident/snash/pkg/term"
	"github.com/LordOfTrident/snash/pkg/prompt"
	"github.com/LordOfTrident/snash/pkg/interpreter"
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
			prompt = e.GenPrompt(os.Getenv("ERR_PROMPT"))
		}

		in := p.Input(prompt)

		err := interpreter.Interpret(e, in, "stdin")
		if err != nil {
			errors.Print(err)
		}

		// Exit the repl if last exit was forced
		if e.Flags.ForcedExit {
			break
		}
	}

	return e.Ex
}
