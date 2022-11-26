package env

import (
	"fmt"
	"os"
	"strings"
	"strconv"

	"github.com/LordOfTrident/snash/internal/symtable"
	"github.com/LordOfTrident/snash/internal/utils"
)

type Env struct {
	Ex int

	Scopes []symtable.Scope

	Flags struct {
		ForcedExit bool
		Echo       bool
	}
}

func New() *Env {
	env := &Env{}

	// Global scope
	env.PushScope()

	// Read the env vars
	for _, v := range os.Environ() {
	    if i := strings.Index(v, "="); i >= 0 {
	        env.Scopes[0].Create(v[:i], v[i + 1:], true)
	    }
	}

	// Defaults
	env.Scopes[0].Create("PROMPT",       "$ ",      false)
	env.Scopes[0].Create("PROMPT_ERROR", "[\\ex] $ ", false)

	env.Scopes[0].Create("USER", os.Getenv("USER"), false)

	return env
}

func (env *Env) PushScope() {
	env.Scopes = append(env.Scopes, symtable.NewScope(len(env.Scopes)))
}

func (env *Env) PopScope() {
	env.Scopes = env.Scopes[:len(env.Scopes) - 1]
}

func (env *Env) Update() (err error) {
	// Update env vars
	if data, err := os.ReadFile("/etc/hostname"); err == nil {
		hostname := strings.Replace(string(data), "\n", "", -1)

		env.Scopes[0].Create("HOSTNAME", hostname, true)
	} else {
		err = fmt.Errorf("Failed to read %v to set %v",
		                 utils.Quote("/etc/hostname"), utils.Quote("$HOSTNAME"))
	}

	if path, err := os.Getwd(); err == nil {
		env.Scopes[0].Create("PWD", path, true)
	} else {
		err = fmt.Errorf("Failed to set %v", utils.Quote("$PWD"))
	}

	if path, err := os.UserHomeDir(); err == nil {
		env.Scopes[0].Create("HOME", path, true)
	} else {
		err = fmt.Errorf("Failed to set %v", utils.Quote("$HOME"))
	}

	return
}

func (env *Env) GenPrompt(prompt string) string {
	// Special escape sequences for the prompt only
	promptSpecials := map[string]string{}

	promptSpecials["\\ex"] = strconv.Itoa(env.Ex)
	promptSpecials["\\w"]  = strings.Replace(env.Scopes[0].Get("PWD"),
	                                         env.Scopes[0].Get("HOME"), "~", -1)
	promptSpecials["\\u"]  = env.Scopes[0].Get("USER")
	promptSpecials["\\h"]  = env.Scopes[0].Get("HOSTNAME")

	// Apply them
	for k, v := range promptSpecials {
		prompt = strings.Replace(prompt, k, v, -1)
	}

	// Special character ignoring markers to ignore escape sequences
	prompt = strings.Replace(prompt, "\\[", "\x01", -1)
	prompt = strings.Replace(prompt, "\\]", "\x02", -1)

	return prompt
}
