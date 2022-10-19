package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{
		Outer: outer,
		store: s,
	}
}

func (st *SymbolTable) Define(ident string) Symbol {
	sym := Symbol{Name: ident, Scope: GlobalScope, Index: st.numDefinitions}
	if st.Outer == nil {
		sym.Scope = GlobalScope
	} else {
		sym.Scope = LocalScope
	}
	st.numDefinitions++
	st.store[ident] = sym
	return sym
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		return obj, ok
	}
	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol{
	symbol := Symbol {Name:name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}
