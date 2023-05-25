package lokum

type Importable interface {
	Import(moduleName string) (interface{}, error)
}

type ModuleGetter interface {
	Get(name string) Importable
}

type ModuleMap struct {
	m map[string]Importable
}

func NewModuleMap() *ModuleMap {
	return &ModuleMap{
		m: make(map[string]Importable),
	}
}

func (m *ModuleMap) Add(name string, module Importable) {
	m.m[name] = module
}

func (m *ModuleMap) AddBuiltinModule(name string, attrs map[string]Object) {
	m.m[name] = &BuiltinModule{Attrs: attrs}
}

func (m *ModuleMap) AddSourceModule(name string, src []byte) {
	m.m[name] = &SourceModule{Src: src}
}

func (m *ModuleMap) Remove(name string) {
	delete(m.m, name)
}

func (m *ModuleMap) Get(name string) Importable {
	return m.m[name]
}

func (m *ModuleMap) GetBuiltinModule(name string) *BuiltinModule {
	mod, _ := m.m[name].(*BuiltinModule)
	return mod
}

func (m *ModuleMap) GetSourceModule(name string) *SourceModule {
	mod, _ := m.m[name].(*SourceModule)
	return mod
}

func (m *ModuleMap) Copy() *ModuleMap {
	c := &ModuleMap{
		m: make(map[string]Importable),
	}
	for name, mod := range m.m {
		c.m[name] = mod
	}
	return c
}

func (m *ModuleMap) Len() int {
	return len(m.m)
}

func (m *ModuleMap) AddMap(o *ModuleMap) {
	for name, mod := range o.m {
		m.m[name] = mod
	}
}

type SourceModule struct {
	Src []byte
}

func (m *SourceModule) Import(_ string) (interface{}, error) {
	return m.Src, nil
}
