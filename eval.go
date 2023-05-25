package lokum

import (
	"context"
	"fmt"
	"strings"
)

func Eval(
	ctx context.Context,
	expr string,
	params map[string]interface{},
) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	script := NewScript([]byte(fmt.Sprintf("__res__ := (%s)", expr)))
	for pk, pv := range params {
		err := script.Add(pk, pv)
		if err != nil {
			return nil, fmt.Errorf("script add: %w", err)
		}
	}
	compiled, err := script.RunContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("script run: %w", err)
	}
	return compiled.Get("__res__").Value(), nil
}
