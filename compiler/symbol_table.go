package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (st *SymbolTable) Define(ident string) Symbol {
	sym := Symbol{Name: ident, Scope: GlobalScope, Index: st.numDefinitions}
	st.numDefinitions++
	st.store[ident] = sym
	return sym
}

func (st *SymbolTable) Resolve(ident string) (Symbol, bool) {
	sym, ok := st.store[ident]
	return sym, ok
}
