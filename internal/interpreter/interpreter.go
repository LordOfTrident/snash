package interpreter

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/node"
	"github.com/LordOfTrident/snash/internal/parser"
	"github.com/LordOfTrident/snash/internal/env"
)

func Interpret(env *env.Env, source, path string) error {
	lastEx := env.Ex
	env.Ex  = 0

	p, err := parser.New(source, path)
	if err != nil {
		env.Ex = 1

		return err
	}

	program, err := p.Parse()
	if err != nil {
		env.Ex = 1

		return err
	}

	// Interpret each statement
	for _, s := range program.List {
		var err error

		switch s := s.(type) {
		case *node.CmdStatement: err = evalCmd(env, s)
		case *node.ExitStatement:
			evalExit(env, s, lastEx)

			return nil

		case *node.EchoStatement: evalEcho(env, s)
		case *node.CdStatement:   err = evalCd(env, s)

		default: err = errors.UnexpectedNode(s)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func evalCmd(env *env.Env, cs *node.CmdStatement) error {
	// Echo the command if echo is enabled
	if env.Flags.Echo {
		fmt.Printf("%v ", cs.Cmd)
	}

	// Read the command arguments
	var args []string
	for _, str := range cs.Args {
		// Echo each argument if echo is enabled
		if env.Flags.Echo {
			fmt.Printf("\"%v \"", str)
		}

		arg, err := evalArg(str)
		if err != nil {
			env.Ex = 1

			return err
		}

		args = append(args, arg)
	}

	// Go to a new line if echo is enabled
	if env.Flags.Echo {
		fmt.Println()
	}

	// If the command does not exist, return exitcode 127
	if _, err := exec.LookPath(cs.Cmd); err != nil {
		env.Ex = 127

		return errors.CmdNotFound(cs.Cmd, cs.NodeToken().Where)
	}

	// Redirect streams and execute the command
	process := exec.Command(cs.Cmd, args...)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	process.Stdin  = os.Stdin

	if err := process.Start(); err != nil {
		panic(err)
	}

	err := process.Wait()
	if exErr, ok := err.(*exec.ExitError); ok {
		env.Ex = exErr.ExitCode()

		return nil
	}

	return nil
}

func evalArg(str string) (string, error) {
	// TODO: implement environment variables in strings

	return str, nil
}

func evalExit(env *env.Env, ex *node.ExitStatement, lastEx int) {
	// Let the environment know that a forced exit happend
	env.Flags.ForcedExit = true

	// If exit does not request an exitcode, use the latest exitcode, otherwise set the exitcode to
	// the requested one
	if ex.HasEx {
		env.Ex = ex.Ex
	} else {
		env.Ex = lastEx
	}
}

func evalEcho(env *env.Env, echo *node.EchoStatement) {
	fmt.Println(echo.Msg)
}

func evalCd(env *env.Env, cd *node.CdStatement) error {
	// Replace the '~' with the home directory path and change the directory
	err := os.Chdir(strings.Replace(cd.Path, "~", os.Getenv("HOME"), -1))
	if err != nil {
		env.Ex = 1

		return errors.FileNotFound(cd.Path, cd.NodeToken().Where)
	}

	return nil
}
