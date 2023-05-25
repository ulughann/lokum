package stdlib

import (
	"github.com/onrirr/lokum"
)

var BuiltinModules = map[string]map[string]lokum.Object{
	"io": fmtModule,
}
