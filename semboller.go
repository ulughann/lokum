package lokum

type SymbolScope string

const (
	ScopeGlobal  SymbolScope = "GLOBAL"
	ScopeLocal   SymbolScope = "LOCAL"
	ScopeBuiltin SymbolScope = "BUILTIN"
	ScopeFree    SymbolScope = "FREE"
)

type Symbol struct {
	Name          string
	Scope         SymbolScope
	Index         int
	LocalAssigned bool
}

type SymbolTable struct {
	parent         *SymbolTable
	block          bool
	store          map[string]*Symbol
	numDefinition  int
	maxDefinition  int
	freeSymbols    []*Symbol
	builtinSymbols []*Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]*Symbol),
	}
}

func (t *SymbolTable) Define(name string) *Symbol {
	symbol := &Symbol{Name: name, Index: t.nextIndex()}
	t.numDefinition++

	if t.Parent(true) == nil {
		symbol.Scope = ScopeGlobal

		if p := t.parent; p != nil {
			for p.parent != nil {
				p = p.parent
			}
			t.numDefinition--
			p.numDefinition++
		}

	} else {
		symbol.Scope = ScopeLocal
	}
	t.store[name] = symbol
	t.updateMaxDefs(symbol.Index + 1)
	return symbol
}

func (t *SymbolTable) DefineBuiltin(index int, name string) *Symbol {
	if t.parent != nil {
		return t.parent.DefineBuiltin(index, name)
	}

	symbol := &Symbol{
		Name:  name,
		Index: index,
		Scope: ScopeBuiltin,
	}
	t.store[name] = symbol
	t.builtinSymbols = append(t.builtinSymbols, symbol)
	return symbol
}

func (t *SymbolTable) Resolve(
	name string,
	recur bool,
) (*Symbol, int, bool) {
	symbol, ok := t.store[name]
	if ok {

		if symbol.Scope != ScopeLocal ||
			symbol.LocalAssigned ||
			recur {
			return symbol, 0, true
		}
	}

	if t.parent == nil {
		return nil, 0, false
	}

	symbol, depth, ok := t.parent.Resolve(name, true)
	if !ok {
		return nil, 0, false
	}
	depth++

	if !t.block && depth > 0 &&
		symbol.Scope != ScopeGlobal &&
		symbol.Scope != ScopeBuiltin {
		return t.defineFree(symbol), depth, true
	}
	return symbol, depth, true
}

func (t *SymbolTable) Fork(block bool) *SymbolTable {
	return &SymbolTable{
		store:  make(map[string]*Symbol),
		parent: t,
		block:  block,
	}
}

func (t *SymbolTable) Parent(skipBlock bool) *SymbolTable {
	if skipBlock && t.block {
		return t.parent.Parent(skipBlock)
	}
	return t.parent
}

func (t *SymbolTable) MaxSymbols() int {
	return t.maxDefinition
}

func (t *SymbolTable) FreeSymbols() []*Symbol {
	return t.freeSymbols
}

func (t *SymbolTable) BuiltinSymbols() []*Symbol {
	if t.parent != nil {
		return t.parent.BuiltinSymbols()
	}
	return t.builtinSymbols
}

func (t *SymbolTable) Names() []string {
	var names []string
	for name := range t.store {
		names = append(names, name)
	}
	return names
}

func (t *SymbolTable) nextIndex() int {
	if t.block {
		return t.parent.nextIndex() + t.numDefinition
	}
	return t.numDefinition
}

func (t *SymbolTable) updateMaxDefs(numDefs int) {
	if numDefs > t.maxDefinition {
		t.maxDefinition = numDefs
	}
	if t.block {
		t.parent.updateMaxDefs(numDefs)
	}
}

func (t *SymbolTable) defineFree(original *Symbol) *Symbol {

	t.freeSymbols = append(t.freeSymbols, original)
	symbol := &Symbol{
		Name:  original.Name,
		Index: len(t.freeSymbols) - 1,
		Scope: ScopeFree,
	}
	t.store[original.Name] = symbol
	return symbol
}
