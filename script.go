package lokum

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/onrirr/lokum/parser"
)

type Script struct {
	variables        map[string]*Variable
	modules          ModuleGetter
	input            []byte
	maxAllocs        int64
	maxConstObjects  int
	enableFileImport bool
	importDir        string
}

func NewScript(input []byte) *Script {
	return &Script{
		variables:       make(map[string]*Variable),
		input:           input,
		maxAllocs:       -1,
		maxConstObjects: -1,
	}
}

func (s *Script) Add(name string, value interface{}) error {
	obj, err := FromInterface(value)
	if err != nil {
		return err
	}
	s.variables[name] = &Variable{
		name:  name,
		value: obj,
	}
	return nil
}

func (s *Script) Remove(name string) bool {
	if _, ok := s.variables[name]; !ok {
		return false
	}
	delete(s.variables, name)
	return true
}

func (s *Script) SetImports(modules ModuleGetter) {
	s.modules = modules
}

func (s *Script) SetImportDir(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	s.importDir = dir
	return nil
}

func (s *Script) SetMaxAllocs(n int64) {
	s.maxAllocs = n
}

func (s *Script) SetMaxConstObjects(n int) {
	s.maxConstObjects = n
}

func (s *Script) EnableFileImport(enable bool) {
	s.enableFileImport = enable
}

func (s *Script) Compile() (*Compiled, error) {
	symbolTable, globals, err := s.prepCompile()
	if err != nil {
		return nil, err
	}

	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile("(main)", -1, len(s.input))
	p := parser.NewParser(srcFile, s.input, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := NewCompiler(srcFile, symbolTable, nil, s.modules, nil)
	c.EnableFileImport(s.enableFileImport)
	c.SetImportDir(s.importDir)
	if err := c.Compile(file); err != nil {
		return nil, err
	}

	globals = globals[:symbolTable.MaxSymbols()+1]

	globalIndexes := make(map[string]int, len(globals))
	for _, name := range symbolTable.Names() {
		symbol, _, _ := symbolTable.Resolve(name, false)
		if symbol.Scope == ScopeGlobal {
			globalIndexes[name] = symbol.Index
		}
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()

	if s.maxConstObjects >= 0 {
		cnt := bytecode.CountObjects()
		if cnt > s.maxConstObjects {
			return nil, fmt.Errorf("sabit obje limiti geçildi: %d", cnt)
		}
	}
	return &Compiled{
		globalIndexes: globalIndexes,
		bytecode:      bytecode,
		globals:       globals,
		maxAllocs:     s.maxAllocs,
	}, nil
}

func (s *Script) Run() (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	err = compiled.Run()
	return
}

func (s *Script) RunContext(
	ctx context.Context,
) (compiled *Compiled, err error) {
	compiled, err = s.Compile()
	if err != nil {
		return
	}
	err = compiled.RunContext(ctx)
	return
}

func (s *Script) prepCompile() (
	symbolTable *SymbolTable,
	globals []Object,
	err error,
) {
	var names []string
	for name := range s.variables {
		names = append(names, name)
	}

	symbolTable = NewSymbolTable()
	for idx, fn := range builtinFuncs {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}

	globals = make([]Object, GlobalsSize)

	for idx, name := range names {
		symbol := symbolTable.Define(name)
		if symbol.Index != idx {
			panic(fmt.Errorf("yanlış sembol dizini: %d != %d",
				idx, symbol.Index))
		}
		globals[symbol.Index] = s.variables[name].value
	}
	return
}

type Compiled struct {
	globalIndexes map[string]int
	bytecode      *Bytecode
	globals       []Object
	maxAllocs     int64
	lock          sync.RWMutex
}

func (c *Compiled) Run() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := NewVM(c.bytecode, c.globals, c.maxAllocs)
	return v.Run()
}

func (c *Compiled) RunContext(ctx context.Context) (err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	v := NewVM(c.bytecode, c.globals, c.maxAllocs)
	ch := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case string:
					ch <- fmt.Errorf(e)
				case error:
					ch <- e
				default:
					ch <- fmt.Errorf("panik: %v", e)
				}
			}
		}()
		ch <- v.Run()
	}()

	select {
	case <-ctx.Done():
		v.Abort()
		<-ch
		err = ctx.Err()
	case err = <-ch:
	}
	return
}

func (c *Compiled) Clone() *Compiled {
	c.lock.RLock()
	defer c.lock.RUnlock()

	clone := &Compiled{
		globalIndexes: c.globalIndexes,
		bytecode:      c.bytecode,
		globals:       make([]Object, len(c.globals)),
		maxAllocs:     c.maxAllocs,
	}

	for idx, g := range c.globals {
		if g != nil {
			clone.globals[idx] = g.Copy()
		}
	}
	return clone
}

func (c *Compiled) IsDefined(name string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	idx, ok := c.globalIndexes[name]
	if !ok {
		return false
	}
	v := c.globals[idx]
	if v == nil {
		return false
	}
	return v != UndefinedValue
}

func (c *Compiled) Get(name string) *Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value := UndefinedValue
	if idx, ok := c.globalIndexes[name]; ok {
		value = c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
	}
	return &Variable{
		name:  name,
		value: value,
	}
}

func (c *Compiled) GetAll() []*Variable {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var vars []*Variable
	for name, idx := range c.globalIndexes {
		value := c.globals[idx]
		if value == nil {
			value = UndefinedValue
		}
		vars = append(vars, &Variable{
			name:  name,
			value: value,
		})
	}
	return vars
}

func (c *Compiled) Set(name string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	obj, err := FromInterface(value)
	if err != nil {
		return err
	}
	idx, ok := c.globalIndexes[name]
	if !ok {
		return fmt.Errorf("'%s' tanımlı değil", name)
	}
	c.globals[idx] = obj
	return nil
}
