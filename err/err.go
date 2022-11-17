package err

import (
	"fmt"

	"github.com/LordOfTrident/snash/token"
)

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
