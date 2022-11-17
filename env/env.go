package env

import (
	"os"
	"strings"
	"strconv"
)

type Env struct {
	Ex int

	Flags struct {
		ForcedExit bool
		Echo       bool
	}
}

func New() *Env {
	return &Env{}
}

func (env *Env) UpdateVars() {
	if data, err := os.ReadFile("/etc/hostname"); err == nil {
		hostname := strings.Replace(string(data), "\n", "", -1)

		os.Setenv("HOSTNAME", hostname)
	}

	if path, err := os.Getwd(); err == nil {
		os.Setenv("PWD", path)
	}
}

func (env *Env) GenPrompt(prompt string) string {
	// Special escape sequences for the prompt only
	promptSpecials := map[string]string{}

	promptSpecials["\\ex"] = strconv.Itoa(env.Ex)
	promptSpecials["\\w"]  = strings.Replace(os.Getenv("PWD"), os.Getenv("HOME"), "~", -1)
	promptSpecials["\\u"]  = os.Getenv("USER")
	promptSpecials["\\h"]  = os.Getenv("HOSTNAME")

	// Apply them
	for k, v := range promptSpecials {
		prompt = strings.Replace(prompt, k, v, -1)
	}

	// Special character ignoring markers to ignore escape sequences
	prompt = strings.Replace(prompt, "\\[", "\x01", -1)
	prompt = strings.Replace(prompt, "\\]", "\x02", -1)

	return prompt
}
