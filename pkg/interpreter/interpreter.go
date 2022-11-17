package interpreter

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/LordOfTrident/snash/pkg/errors"
	"github.com/LordOfTrident/snash/pkg/node"
	"github.com/LordOfTrident/snash/pkg/parser"
	"github.com/LordOfTrident/snash/pkg/env"
)

func exitStatusToExitCode(status string) int {
	// Golang commands error if the exitcode is not 0, but they dont return the exitcode itself,
	// so we have to get the exitcode from the error message (ugly, but whatever)

	// The error message format is 'exit status <n>' where '<n>' is the exitcode,
	// so we can just chop off the prefix 'exit status '
	msg := "exit status "

	if status[:len(msg)] == msg {
		status = status[len(msg):]
	}

	ex, err := strconv.Atoi(status)
	if err != nil {
		panic(err)
	}

	return ex
}

func cmdNotFound(cmd string, node node.Node) error {
	return errors.New(node.NodeToken().Where, "Command '%v' not found", cmd)
}

func fileNotFound(path string, node node.Node) error {
	return errors.New(node.NodeToken().Where, "File/directory '%v' not found", path)
}

func unexpected(node node.Node) error {
	return errors.New(node.NodeToken().Where, "Unexpected %v", node.NodeToken())
}

func Interpret(env *env.Env, source, path string) error {
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
			evalExit(env, s)

			return nil

		case *node.EchoStatement: evalEcho(env, s)
		case *node.CdStatement:   err = evalCd(env, s)

		default: err = unexpected(s) // TODO: this
		}

		if err != nil {
			return err
		}
	}

	env.Ex = 0

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

		return cmdNotFound(cs.Cmd, cs)
	}

	// Redirect streams and execute the command
	process := exec.Command(cs.Cmd, args...)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	process.Stdin  = os.Stdin

	err := process.Start()
	if err == nil {
		err = process.Wait()
	}

	if werr, ok := err.(*exec.ExitError); ok {
		// Parse the error message to get the command exitcode
		env.Ex = exitStatusToExitCode(werr.Error())

		return nil
	}

	return nil
}

func evalArg(str string) (string, error) {
	// TODO: implement environment variables in strings

	return str, nil
}

func evalExit(env *env.Env, ex *node.ExitStatement) {
	// Let the environment know that a forced exit happend
	env.Flags.ForcedExit = true

	// If exit does not request an exitcode, use the latest exitcode, otherwise set the exitcode to
	// the requested one
	if ex.HasEx {
		env.Ex = ex.Ex
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

		return fileNotFound(cd.Path, cd)
	}

	return nil
}
