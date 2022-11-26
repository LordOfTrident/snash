package repl

import (
	"fmt"

	"github.com/LordOfTrident/snash/pkg/term"
	"github.com/LordOfTrident/snash/pkg/prompt"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/env"
	"github.com/LordOfTrident/snash/internal/config"
	"github.com/LordOfTrident/snash/internal/evaluator"
	"github.com/LordOfTrident/snash/internal/highlighter"
)

func REPL(env *env.Env) int {
	term.OnCtrlC(func() {})
	term.SendResizeEvents()

	history, _ := prompt.LoadHistory(config.HistoryFile)

	h := highlighter.New(env)
	p := prompt.New(history, h)

	p.Flags.Interactive        = *config.Interactive
	p.Flags.ShowPossibleErrors = *config.ShowPossibleErrors
	p.Flags.SyntaxHighlighting = *config.SyntaxHighlighting

	for {
		env.Update()

		// Generate a prompt
		var prompt string
		if env.Ex == 0 {
			prompt = env.GenPrompt(env.Scopes[0].Get("PROMPT"))
		} else {
			prompt = env.GenPrompt(env.Scopes[0].Get("PROMPT_ERROR"))
		}

		in := p.Input(prompt)

		err := evaluator.Eval(env, in, "stdin")
		if err != nil {
			highlighter.PrintError(err)
		}

		// Exit the repl if last exit was forced
		if env.Flags.ForcedExit {
			break
		}
	}

	if err := p.History.SaveToFile(config.HistoryFile); err != nil {
		highlighter.PrintError(fmt.Errorf("Could not save history file %v",
		                                  utils.Quote(config.HistoryFile)))
	}

	return env.Ex
}
