package interpreter

import (
	"os"
	"os/exec"
	"fmt"
	"strconv"

	"github.com/LordOfTrident/snash/compilerError"
	"github.com/LordOfTrident/snash/node"
	"github.com/LordOfTrident/snash/parser"
)

func Interpret(source, path string) (int, bool, error) {
	p, err := parser.New(source, path)
	if err != nil {
		return 1, true, err
	}

	program, err := p.Parse()
	if err != nil {
		return 1, true, err
	}

	ex := 0
	for _, s := range program.List {
		var err error

		switch s := s.(type) {
		case *node.CmdStatement:  ex, err = evalCmd(s)
		case *node.ExitStatement: return evalExit(s), true, nil

		default: err = compilerError.UnexpectedNode(s)
		}

		if err != nil {
			return 1, true, err
		}
	}

	return ex, false, nil
}

func evalCmd(cs *node.CmdStatement) (int, error) {
	var args []string
	for _, str := range cs.Args {
		arg, err := evalArg(str)
		if err != nil {
			return 1, err
		}

		args = append(args, arg)
	}

	if _, err := exec.LookPath(cs.Cmd); err != nil {
		return 127, fmt.Errorf("Command '%v' not found", cs.Cmd)
	}

	process := exec.Command(cs.Cmd, args...)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout

	err := process.Run()
	if werr, ok := err.(*exec.ExitError); ok {
		ex, err := strconv.Atoi(werr.Error())
		if err != nil {
			panic(err)
		}

		return ex, nil
	}

	return 0, nil
}

func evalArg(str string) (string, error) {
	// TODO: implement environment variables in strings

	return str, nil
}

func evalExit(ex *node.ExitStatement) int {
	return ex.Exitcode
}
