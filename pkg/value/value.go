package value

import (
	"fmt"
	"strconv"
)

type Value interface {
	ValueTypeToString() string
	ValueToString()     string
}

type StringValue struct {
	Data string
}

func (s *StringValue) ValueTypeToString() string {
	return "string"
}

func (s *StringValue) ValueToString() string {
	return s.Data
}

type IntegerValue struct {
	Data int
}

func (i *IntegerValue) ValueTypeToString() string {
	return "integer"
}

func (i *IntegerValue) ValueToString() string {
	return strconv.Itoa(s.Data)
}
