package repl

import (
	"os"

	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/env"
	"github.com/LordOfTrident/snash/pkg/prompt"
	"github.com/LordOfTrident/snash/pkg/interpreter"
)

func REPL(interactive bool) int {
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

		in := p.ReadLine(prompt, interactive)

		err := interpreter.Interpret(env, in, "stdin")
		if err != nil {
			errors.Print(err)
		}

		// Exit the repl if last exit was forced
		if env.Flags.ForcedExit {
			break
		}
	}

	return env.Ex
}
