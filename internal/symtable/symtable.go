package symtable

// TODO: Use interfaces to allow variables AND functions in symbol tables,
//       function bodies are gonna be just nodes

type Entry struct {
	Value  string
	Export bool
}

func NewEntry(value string, export bool) Entry {
	return Entry{Value: value, Export: export}
}

type Scope struct {
	SymTable map[string]Entry
	Level    int
}

func NewScope(level int) Scope {
	return Scope{SymTable: make(map[string]Entry), Level: level}
}

func (s *Scope) Create(name, value string, export bool) {
	s.SymTable[name] = NewEntry(value, export)
}

func (s *Scope) Export(name string, export bool) {
	if entry, ok := s.SymTable[name]; ok {
		entry.Export = export;

		s.SymTable[name] = entry
	}
}

func (s *Scope) Set(name, value string) {
	if entry, ok := s.SymTable[name]; ok {
		entry.Value = value;

		s.SymTable[name] = entry
	}
}

func (s *Scope) Exists(name string) bool {
	_, ok := s.SymTable[name]

	return ok
}

func (s *Scope) Get(name string) string {
	return s.SymTable[name].Value
}

func (s *Scope) Unset(name string) {
	delete(s.SymTable, name)
}
