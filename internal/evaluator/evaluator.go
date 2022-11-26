package evaluator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/LordOfTrident/snash/pkg/term"

	"github.com/LordOfTrident/snash/internal/utils"
	"github.com/LordOfTrident/snash/internal/config"
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
		if _, err := evalStatement(env, s); err != nil {
			return err
		}
	}

	return nil
}

func evalStatement(env *env.Env, s node.Statement) (ex int, err error) {
	switch s := s.(type) {
	case *node.LetStatement:    err = evalLet(env, s)
	case *node.AssignStatement: err = evalAssign(env, s)
	case *node.ExportStatement: err = evalExport(env, s)

	case *node.CmdStatement:  ex, err = evalCmd(env, s)
	case *node.CdStatement:   err     = evalCd(env, s)
	case *node.ExitStatement: ex      = evalExit(env, s)
	case *node.HelpStatement:           evalHelp(env, s)
	case *node.EchoStatement:           evalEcho(env, s)

	case *node.BinOpStatement: ex, err = evalBinOp(env, s)

	default: err = errors.UnexpectedNode(s)
	}

	if err != nil && ex == 0 {
		ex = 1
	}

	env.Ex = ex

	return
}

func evalLet(env *env.Env, let *node.LetStatement) error {
	for i, ch := range let.Name {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return errors.New(let.NodeToken().Where,
			                     "Unexpected character %v at %v in variable identifier",
			                     utils.Quote(string(ch)), i)
		}
	}

	env.Scopes[0].Create(let.Name, let.Value, false)

	return nil
}

func evalAssign(env *env.Env, as *node.AssignStatement) error {
	if !env.Scopes[0].Exists(as.Name) {
		return errors.VarNotFound(as.Name, as.NodeToken().Where)
	}

	env.Scopes[0].Set(as.Name, as.Value)

	return nil
}

func evalExport(env *env.Env, export *node.ExportStatement) error {
	if !env.Scopes[0].Exists(export.Name) {
		return errors.VarNotFound(export.Name, export.NodeToken().Where)
	}

	env.Scopes[0].Export(export.Name, true)

	return nil
}

func evalBinOp(env *env.Env, bin *node.BinOpStatement) (int, error) {
	switch bin.NodeToken().Type {
	case token.Or:  return evalOrBinOp(env, bin)
	case token.And: return evalAndBinOp(env, bin)

	default: return 1, errors.UnexpectedNode(bin)
	}
}

func evalOrBinOp(env *env.Env, bin *node.BinOpStatement) (int, error) {
	ex, err := evalStatement(env, bin.Left)
	if err != nil {
		return 1, err
	}

	// Only run the second command if the first one failed
	if env.Ex != 0 {
		ex, err := evalStatement(env, bin.Right)
		if err != nil {
			return 1, err
		}

		return ex, nil
	}

	return ex, nil
}

func evalAndBinOp(env *env.Env, bin *node.BinOpStatement) (int, error) {
	ex, err := evalStatement(env, bin.Left)
	if err != nil {
		return 1, err
	}

	// Only run the second command if the first one didnt error
	if env.Ex == 0 {
		ex, err := evalStatement(env, bin.Right)
		if err != nil {
			return 1, err
		}

		return ex, nil
	}

	return ex, nil
}

func evalCmd(env *env.Env, cs *node.CmdStatement) (int, error) {
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

		args = append(args, str)
	}

	// Go to a new line if echo is enabled
	if env.Flags.Echo {
		fmt.Println()
	}

	// If the command does not exist, return exitcode 127
	if _, err := exec.LookPath(cs.Cmd); err != nil {
		return 127, errors.CmdNotFound(cs.Cmd, cs.NodeToken().Where)
	}

	// Redirect streams and execute the command
	process := exec.Command(cs.Cmd, args...)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	process.Stdin  = os.Stdin

	process.Env = []string{}
	for i, v := range env.Scopes[0].SymTable {
		if v.Export {
			process.Env = append(process.Env, i + "=" + v.Value)
		}
	}

	if err := process.Start(); err != nil {
		panic(err)
	}

	err := process.Wait()
	if exErr, ok := err.(*exec.ExitError); ok {
		return exErr.ExitCode(), nil
	}

	return 0, nil
}

func evalExit(env *env.Env, ex *node.ExitStatement) int {
	// Let the environment know that a forced exit happend
	env.Flags.ForcedExit = true

	// If exit does not request an exitcode, use the latest exitcode, otherwise set the exitcode to
	// the requested one
	if ex.HasEx {
		return ex.Ex
	}

	return env.Ex
}

func keywordHighlight(keyword string) string {
	return term.AttrBold + term.AttrBrightBlue + keyword + term.AttrReset
}

func evalHelp(env *env.Env, echo *node.HelpStatement) {
	fmt.Println(term.AttrUnderline + config.GithubLink + term.AttrReset)

	fmt.Println("\n  " + term.AttrGreen + term.AttrItalics +
	            "A shell for Unix and Linux systems" + term.AttrReset +"\n")

	fmt.Printf("%v help\nversion " + term.AttrCyan + "%v.%v.%v" + term.AttrReset + "\n",
	           config.AppName, config.VersionMajor, config.VersionMinor, config.VersionPatch)

	fmt.Println("\nBuilt-in commands:")
	fmt.Printf("  %v           Show this message\n",                 keywordHighlight("help"))
	fmt.Printf("  %v [str...]  Output a string\n",                   keywordHighlight("echo"))
	fmt.Printf("  %v [int]     Exit the process with an exitcode\n", keywordHighlight("exit"))
	fmt.Printf("  %v [path]    Change the current directory\n",      keywordHighlight("cd  "))
}

func evalEcho(env *env.Env, echo *node.EchoStatement) {
	fmt.Println(echo.Msg)
}

func evalCd(env *env.Env, cd *node.CdStatement) error {
	// Replace the '~' with the home directory path and change the directory
	err := os.Chdir(strings.Replace(cd.Path, "~", os.Getenv("HOME"), -1))
	if err != nil {
		return errors.FileNotFound(cd.Path, cd.NodeToken().Where)
	}

	return nil
}
