package stdlib

//go:generate go run gensrcmods.go

import (
	"github.com/onrirr/lokum"
)

func AllModuleNames() []string {
	var names []string
	for name := range BuiltinModules {
		names = append(names, name)
	}
	for name := range SourceModules {
		names = append(names, name)
	}
	return names
}

func GetModuleMap(names ...string) *lokum.ModuleMap {
	modules := lokum.NewModuleMap()
	for _, name := range names {
		if mod := BuiltinModules[name]; mod != nil {
			modules.AddBuiltinModule(name, mod)
		}
		if mod := SourceModules[name]; mod != "" {
			modules.AddSourceModule(name, []byte(mod))
		}
	}
	return modules
}
