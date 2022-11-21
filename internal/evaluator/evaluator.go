package evaluator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LordOfTrident/snash/internal/errors"
	"github.com/LordOfTrident/snash/internal/token"
	"github.com/LordOfTrident/snash/internal/node"
	"github.com/LordOfTrident/snash/internal/parser"
	"github.com/LordOfTrident/snash/internal/env"
)

func Eval(env *env.Env, source, path string) error {
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

	return evalStatements(env, program)
}

func evalStatements(env *env.Env, statements node.Statements) error {
	for _, s := range statements.List {
		if err := evalStatement(env, s); err != nil {
			return err
		}
	}

	return nil
}

func evalStatement(env *env.Env, s node.Statement) error {
	switch s := s.(type) {
	case *node.CmdStatement:  return evalCmd(env, s)
	case *node.CdStatement:   return evalCd(env, s)
	case *node.ExitStatement: evalExit(env, s)
	case *node.EchoStatement: evalEcho(env, s)

	case *node.BinOpStatement: return evalBinOp(env, s)

	default: return errors.UnexpectedNode(s)
	}

	return nil
}

func evalBinOp(env *env.Env, bin *node.BinOpStatement) error {
	switch bin.NodeToken().Type {
	case token.Or:  return evalOrBinOp(env, bin)
	case token.And: return evalAndBinOp(env, bin)

	default: return errors.UnexpectedNode(bin)
	}
}

func evalOrBinOp(env *env.Env, bin *node.BinOpStatement) error {
	err := evalStatement(env, bin.Left)
	if err != nil {
		return err
	}

	if env.Ex != 0 {
		return evalStatement(env, bin.Right)
	}

	return nil
}

func evalAndBinOp(env *env.Env, bin *node.BinOpStatement) error {
	err := evalStatement(env, bin.Left)
	if err != nil {
		return err
	}

	if env.Ex == 0 {
		return evalStatement(env, bin.Right)
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
	} else {
		env.Ex = 0
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

	env.Ex = 0
}

func evalCd(env *env.Env, cd *node.CdStatement) error {
	// Replace the '~' with the home directory path and change the directory
	err := os.Chdir(strings.Replace(cd.Path, "~", os.Getenv("HOME"), -1))
	if err != nil {
		env.Ex = 1

		return errors.FileNotFound(cd.Path, cd.NodeToken().Where)
	} else {
		env.Ex = 0
	}

	return nil
}
