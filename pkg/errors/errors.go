package errors

import (
	"fmt"
	"os"

	"github.com/LordOfTrident/snash/pkg/token"
	"github.com/LordOfTrident/snash/pkg/attr"
)

func Print(err error) {
	fmt.Fprintf(os.Stderr, "%v%vError:%v %v\n", attr.Bold, attr.BrightRed, attr.Reset, err.Error())
}

type Error struct {
	Where token.Where
	Msg   string
}

func (err Error) Error() string {
	return fmt.Sprintf("%v: %v", err.Where, err.Msg)
}

func New(where token.Where, format string, args... interface{}) error {
	return Error{Where: where, Msg: fmt.Sprintf(format, args...)}
}
