package symboltable

type Type int
const (
	String = iota
	Func
)

type Entry struct {
	Name, Value string
	Type        Type
	Export      bool
}
