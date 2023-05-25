package stdlib

import (
	"github.com/onrirr/lokum"
)

func wrapError(err error) lokum.Object {
	if err == nil {
		return lokum.TrueValue
	}
	return &lokum.Error{Value: &lokum.String{Value: err.Error()}}
}
