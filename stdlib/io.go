package stdlib

import (
	"fmt"

	"github.com/onrirr/lokum"
)

var fmtModule = map[string]lokum.Object{
	"yazdırnf": &lokum.UserFunction{Name: "yazdırnf", Value: fmtPrint},
	"yazdırf":  &lokum.UserFunction{Name: "yazdırf", Value: fmtPrintf},
	"yazdır":   &lokum.UserFunction{Name: "yazdır", Value: fmtPrintln},
	"sprintf":  &lokum.UserFunction{Name: "sprintf", Value: fmtSprintf},
}

func fmtPrint(args ...lokum.Object) (ret lokum.Object, err error) {
	printArgs, err := getPrintArgs(args...)
	if err != nil {
		return nil, err
	}
	_, _ = fmt.Print(printArgs...)
	return nil, nil
}

func fmtPrintf(args ...lokum.Object) (ret lokum.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, lokum.ErrWrongNumArguments
	}

	format, ok := args[0].(*lokum.String)
	if !ok {
		return nil, lokum.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	if numArgs == 1 {
		fmt.Print(format)
		return nil, nil
	}

	s, err := lokum.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}
	fmt.Print(s)
	return nil, nil
}

func fmtPrintln(args ...lokum.Object) (ret lokum.Object, err error) {
	printArgs, err := getPrintArgs(args...)
	if err != nil {
		return nil, err
	}
	printArgs = append(printArgs, "\n")
	_, _ = fmt.Print(printArgs...)
	return nil, nil
}

func fmtSprintf(args ...lokum.Object) (ret lokum.Object, err error) {
	numArgs := len(args)
	if numArgs == 0 {
		return nil, lokum.ErrWrongNumArguments
	}

	format, ok := args[0].(*lokum.String)
	if !ok {
		return nil, lokum.ErrInvalidArgumentType{
			Name:     "format",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}
	if numArgs == 1 {

		return format, nil
	}
	s, err := lokum.Format(format.Value, args[1:]...)
	if err != nil {
		return nil, err
	}
	return &lokum.String{Value: s}, nil
}

func getPrintArgs(args ...lokum.Object) ([]interface{}, error) {
	var printArgs []interface{}
	l := 0
	for _, arg := range args {
		s, _ := lokum.ToString(arg)
		slen := len(s)

		if l+slen > lokum.MaxStringLen {
			return nil, lokum.ErrStringLimit
		}
		l += slen
		printArgs = append(printArgs, s)
	}
	return printArgs, nil
}
