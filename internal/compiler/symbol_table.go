package compiler

type SymbolScope string

const (
	// GlobalScope identifies a variable that is global
	GlobalScope SymbolScope = "GLOBAL"
	// LocalScope identifies a variable in a local scope
	LocalScope SymbolScope = "LOCAL"
	// BuiltinScope identified a builtin function
	BuiltinScope SymbolScope = "BUILTIN"
)

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Symbol holds all the necessary information about a symbol encountered in the code
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable associates strings with Symbols in its `store` and keeps track
// of the number of definitions it has
type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

// NewSymbolTable creates a new symbol table and returns a pointer to it.
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// Define defines a new symbol in the symbol table with the given name.
// It returns the created symbol.
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: GlobalScope}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

// Resolve looks up a symbol by name in the symbol table and returns the corresponding symbol object and a boolean indicating if the symbol was found.
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok := s.Outer.Resolve(name)
		return obj, ok
	}
	return obj, ok
}
