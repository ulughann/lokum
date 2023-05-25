package lokum

import "fmt"

var builtinFuncs = []*BuiltinFunction{
	{
		Name: "yazdır",
		Value: func(args ...Object) (Object, error) {
			for _, arg := range args {
				fmt.Println(arg)
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "uzunluk",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Int{Value: int64(len(arg.Value))}, nil
			case *ImmutableArray:
				return &Int{Value: int64(len(arg.Value))}, nil
			case *String:
				return &Int{Value: int64(len(arg.Value))}, nil
			case *Bytes:
				return &Int{Value: int64(len(arg.Value))}, nil
			case *Map:
				return &Int{Value: int64(len(arg.Value))}, nil
			case *ImmutableMap:
				return &Int{Value: int64(len(arg.Value))}, nil
			default:
				return nil, ErrInvalidArgumentType{
					Name:     "first",
					Expected: "array/string/bytes/map",
					Found:    arg.TypeName(),
				}
			}
		},
	},
	{
		Name: "kopyala",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			return args[0].Copy(), nil
		},
	},
	{
		Name: "ekle",
		Value: func(args ...Object) (Object, error) {
			if len(args) < 2 {
				return nil, ErrWrongNumArguments
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Array{Value: append(arg.Value, args[1:]...)}, nil
			case *ImmutableArray:
				return &Array{Value: append(arg.Value, args[1:]...)}, nil
			default:
				return nil, ErrInvalidArgumentType{
					Name:     "first",
					Expected: "array",
					Found:    arg.TypeName(),
				}
			}
		},
	},
	{
		Name: "sil",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if argsLen != 2 {
				return nil, ErrWrongNumArguments
			}
			switch arg := args[0].(type) {
			case *Map:
				if key, ok := args[1].(*String); ok {
					delete(arg.Value, key.Value)
					return UndefinedValue, nil
				}
				return nil, ErrInvalidArgumentType{
					Name:     "second",
					Expected: "string",
					Found:    args[1].TypeName(),
				}
			default:
				return nil, ErrInvalidArgumentType{
					Name:     "first",
					Expected: "map",
					Found:    arg.TypeName(),
				}
			}
		},
	},
	{
		Name: "birleştir",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if argsLen == 0 {
				return nil, ErrWrongNumArguments
			}

			array, ok := args[0].(*Array)
			if !ok {
				return nil, ErrInvalidArgumentType{
					Name:     "first",
					Expected: "array",
					Found:    args[0].TypeName(),
				}
			}
			arrayLen := len(array.Value)

			var startIdx int
			if argsLen > 1 {
				arg1, ok := args[1].(*Int)
				if !ok {
					return nil, ErrInvalidArgumentType{
						Name:     "second",
						Expected: "int",
						Found:    args[1].TypeName(),
					}
				}
				startIdx = int(arg1.Value)
				if startIdx < 0 || startIdx > arrayLen {
					return nil, ErrIndexOutOfBounds
				}
			}

			delCount := len(array.Value)
			if argsLen > 2 {
				arg2, ok := args[2].(*Int)
				if !ok {
					return nil, ErrInvalidArgumentType{
						Name:     "third",
						Expected: "int",
						Found:    args[2].TypeName(),
					}
				}
				delCount = int(arg2.Value)
				if delCount < 0 {
					return nil, ErrIndexOutOfBounds
				}
			}

			if startIdx+delCount > arrayLen {
				delCount = arrayLen - startIdx
			}

			endIdx := startIdx + delCount
			deleted := append([]Object{}, array.Value[startIdx:endIdx]...)

			head := array.Value[:startIdx]
			var items []Object
			if argsLen > 3 {
				items = make([]Object, 0, argsLen-3)
				for i := 3; i < argsLen; i++ {
					items = append(items, args[i])
				}
			}
			items = append(items, array.Value[endIdx:]...)
			array.Value = append(head, items...)

			return &Array{Value: deleted}, nil
		},
	},
	{
		Name: "yazı",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if !(argsLen == 1 || argsLen == 2) {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*String); ok {
				return args[0], nil
			}
			v, ok := ToString(args[0])
			if ok {
				if len(v) > MaxStringLen {
					return nil, ErrStringLimit
				}
				return &String{Value: v}, nil
			}
			if argsLen == 2 {
				return args[1], nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "sayı",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if !(argsLen == 1 || argsLen == 2) {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Int); ok {
				return args[0], nil
			}
			v, ok := ToInt64(args[0])
			if ok {
				return &Int{Value: v}, nil
			}
			if argsLen == 2 {
				return args[1], nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "mantıksal",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Bool); ok {
				return args[0], nil
			}
			v, ok := ToBool(args[0])
			if ok {
				if v {
					return TrueValue, nil
				}
				return FalseValue, nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "float",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if !(argsLen == 1 || argsLen == 2) {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Float); ok {
				return args[0], nil
			}
			v, ok := ToFloat64(args[0])
			if ok {
				return &Float{Value: v}, nil
			}
			if argsLen == 2 {
				return args[1], nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "karakter",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if !(argsLen == 1 || argsLen == 2) {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Char); ok {
				return args[0], nil
			}
			v, ok := ToRune(args[0])
			if ok {
				return &Char{Value: v}, nil
			}
			if argsLen == 2 {
				return args[1], nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "bytes",
		Value: func(args ...Object) (Object, error) {
			argsLen := len(args)
			if !(argsLen == 1 || argsLen == 2) {
				return nil, ErrWrongNumArguments
			}

			if n, ok := args[0].(*Int); ok {
				if n.Value > int64(MaxBytesLen) {
					return nil, ErrBytesLimit
				}
				return &Bytes{Value: make([]byte, int(n.Value))}, nil
			}
			v, ok := ToByteSlice(args[0])
			if ok {
				if len(v) > MaxBytesLen {
					return nil, ErrBytesLimit
				}
				return &Bytes{Value: v}, nil
			}
			if argsLen == 2 {
				return args[1], nil
			}
			return UndefinedValue, nil
		},
	},
	{
		Name: "sayı_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Int); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "float_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Float); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "yazı_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*String); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "mantıksal_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Bool); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "liste_mi",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Array); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "harita_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if _, ok := args[0].(*Map); ok {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "tanımsız_mı",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			if args[0] == UndefinedValue {
				return TrueValue, nil
			}
			return FalseValue, nil
		},
	},
	{
		Name: "sınıf",
		Value: func(args ...Object) (Object, error) {
			if len(args) != 1 {
				return nil, ErrWrongNumArguments
			}
			return &String{Value: args[0].TypeName()}, nil
		},
	},
	{
		Name: "f",
		Value: func(args ...Object) (Object, error) {
			numArgs := len(args)
			if numArgs == 0 {
				return nil, ErrWrongNumArguments
			}
			format, ok := args[0].(*String)
			if !ok {
				return nil, ErrInvalidArgumentType{
					Name:     "format",
					Expected: "string",
					Found:    args[0].TypeName(),
				}
			}
			if numArgs == 1 {

				return format, nil
			}
			s, err := Format(format.Value, args[1:]...)
			if err != nil {
				return nil, err
			}
			return &String{Value: s}, nil
		},
	},
	{
		Name: "aralık",
		Value: func(args ...Object) (Object, error) {
			numArgs := len(args)
			if numArgs < 2 || numArgs > 3 {
				return nil, ErrWrongNumArguments
			}
			var start, stop, step *Int

			for i, arg := range args {
				v, ok := args[i].(*Int)
				if !ok {
					var name string
					switch i {
					case 0:
						name = "start"
					case 1:
						name = "stop"
					case 2:
						name = "step"
					}

					return nil, ErrInvalidArgumentType{
						Name:     name,
						Expected: "int",
						Found:    arg.TypeName(),
					}
				}
				if i == 2 && v.Value <= 0 {
					return nil, ErrInvalidRangeStep
				}
				switch i {
				case 0:
					start = v
				case 1:
					stop = v
				case 2:
					step = v
				}
			}

			if step == nil {
				step = &Int{Value: int64(1)}
			}

			return buildRange(start.Value, stop.Value, step.Value), nil
		},
	},
}

func GetAllBuiltinFunctions() []*BuiltinFunction {
	return append([]*BuiltinFunction{}, builtinFuncs...)
}

func buildRange(start, stop, step int64) *Array {
	array := &Array{}
	if start <= stop {
		for i := start; i < stop; i += step {
			array.Value = append(array.Value, &Int{
				Value: i,
			})
		}
	} else {
		for i := start; i > stop; i -= step {
			array.Value = append(array.Value, &Int{
				Value: i,
			})
		}
	}
	return array
}
