package lokum

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/onrirr/lokum/parser"
	"github.com/onrirr/lokum/token"
)

var (
	TrueValue Object = &Bool{value: true}

	FalseValue Object = &Bool{value: false}

	UndefinedValue Object = &Undefined{}
)

type Object interface {
	TypeName() string

	String() string

	BinaryOp(op token.Token, rhs Object) (Object, error)

	IsFalsy() bool

	Equals(another Object) bool

	Copy() Object

	IndexGet(index Object) (value Object, err error)

	IndexSet(index, value Object) error

	Iterate() Iterator

	CanIterate() bool

	Call(args ...Object) (ret Object, err error)

	CanCall() bool
}

type ObjectImpl struct {
}

func (o *ObjectImpl) TypeName() string {
	panic(ErrNotImplemented)
}

func (o *ObjectImpl) String() string {
	panic(ErrNotImplemented)
}

func (o *ObjectImpl) BinaryOp(_ token.Token, _ Object) (Object, error) {
	return nil, ErrInvalidOperator
}

func (o *ObjectImpl) Copy() Object {
	return nil
}

func (o *ObjectImpl) IsFalsy() bool {
	return false
}

func (o *ObjectImpl) Equals(x Object) bool {
	return o == x
}

func (o *ObjectImpl) IndexGet(_ Object) (res Object, err error) {
	return nil, ErrNotIndexable
}

func (o *ObjectImpl) IndexSet(_, _ Object) (err error) {
	return ErrNotIndexAssignable
}

func (o *ObjectImpl) Iterate() Iterator {
	return nil
}

func (o *ObjectImpl) CanIterate() bool {
	return false
}

func (o *ObjectImpl) Call(_ ...Object) (ret Object, err error) {
	return nil, nil
}

func (o *ObjectImpl) CanCall() bool {
	return false
}

type Array struct {
	ObjectImpl
	Value []Object
}

func (o *Array) TypeName() string {
	return "array"
}

func (o *Array) String() string {
	var elements []string
	for _, e := range o.Value {
		elements = append(elements, e.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (o *Array) BinaryOp(op token.Token, rhs Object) (Object, error) {
	if rhs, ok := rhs.(*Array); ok {
		switch op {
		case token.Add:
			if len(rhs.Value) == 0 {
				return o, nil
			}
			return &Array{Value: append(o.Value, rhs.Value...)}, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Array) Copy() Object {
	var c []Object
	for _, elem := range o.Value {
		c = append(c, elem.Copy())
	}
	return &Array{Value: c}
}

func (o *Array) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *Array) Equals(x Object) bool {
	var xVal []Object
	switch x := x.(type) {
	case *Array:
		xVal = x.Value
	case *ImmutableArray:
		xVal = x.Value
	default:
		return false
	}
	if len(o.Value) != len(xVal) {
		return false
	}
	for i, e := range o.Value {
		if !e.Equals(xVal[i]) {
			return false
		}
	}
	return true
}

func (o *Array) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	idxVal := int(intIdx.Value)
	if idxVal < 0 || idxVal >= len(o.Value) {
		res = UndefinedValue
		return
	}
	res = o.Value[idxVal]
	return
}

func (o *Array) IndexSet(index, value Object) (err error) {
	intIdx, ok := ToInt(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	if intIdx < 0 || intIdx >= len(o.Value) {
		err = ErrIndexOutOfBounds
		return
	}
	o.Value[intIdx] = value
	return nil
}

func (o *Array) Iterate() Iterator {
	return &ArrayIterator{
		v: o.Value,
		l: len(o.Value),
	}
}

func (o *Array) CanIterate() bool {
	return true
}

type Bool struct {
	ObjectImpl

	value bool
}

func (o *Bool) String() string {
	if o.value {
		return "true"
	}

	return "false"
}

func (o *Bool) TypeName() string {
	return "bool"
}

func (o *Bool) Copy() Object {
	return o
}

func (o *Bool) IsFalsy() bool {
	return !o.value
}

func (o *Bool) Equals(x Object) bool {
	return o == x
}

func (o *Bool) GobDecode(b []byte) (err error) {
	o.value = b[0] == 1
	return
}

func (o *Bool) GobEncode() (b []byte, err error) {
	if o.value {
		b = []byte{1}
	} else {
		b = []byte{0}
	}
	return
}

type BuiltinFunction struct {
	ObjectImpl
	Name  string
	Value CallableFunc
}

func (o *BuiltinFunction) TypeName() string {
	return "builtin-function:" + o.Name
}

func (o *BuiltinFunction) String() string {
	return "<builtin-function>"
}

func (o *BuiltinFunction) Copy() Object {
	return &BuiltinFunction{Value: o.Value}
}

func (o *BuiltinFunction) Equals(_ Object) bool {
	return false
}

func (o *BuiltinFunction) Call(args ...Object) (Object, error) {
	return o.Value(args...)
}

func (o *BuiltinFunction) CanCall() bool {
	return true
}

type BuiltinModule struct {
	Attrs map[string]Object
}

func (m *BuiltinModule) Import(moduleName string) (interface{}, error) {
	return m.AsImmutableMap(moduleName), nil
}

func (m *BuiltinModule) AsImmutableMap(moduleName string) *ImmutableMap {
	attrs := make(map[string]Object, len(m.Attrs))
	for k, v := range m.Attrs {
		attrs[k] = v.Copy()
	}
	attrs["__module_name__"] = &String{Value: moduleName}
	return &ImmutableMap{Value: attrs}
}

type Bytes struct {
	ObjectImpl
	Value []byte
}

func (o *Bytes) String() string {
	return string(o.Value)
}

func (o *Bytes) TypeName() string {
	return "bytes"
}

func (o *Bytes) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch op {
	case token.Add:
		switch rhs := rhs.(type) {
		case *Bytes:
			if len(o.Value)+len(rhs.Value) > MaxBytesLen {
				return nil, ErrBytesLimit
			}
			return &Bytes{Value: append(o.Value, rhs.Value...)}, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Bytes) Copy() Object {
	return &Bytes{Value: append([]byte{}, o.Value...)}
}

func (o *Bytes) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *Bytes) Equals(x Object) bool {
	t, ok := x.(*Bytes)
	if !ok {
		return false
	}
	return bytes.Equal(o.Value, t.Value)
}

func (o *Bytes) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	idxVal := int(intIdx.Value)
	if idxVal < 0 || idxVal >= len(o.Value) {
		res = UndefinedValue
		return
	}
	res = &Int{Value: int64(o.Value[idxVal])}
	return
}

func (o *Bytes) Iterate() Iterator {
	return &BytesIterator{
		v: o.Value,
		l: len(o.Value),
	}
}

func (o *Bytes) CanIterate() bool {
	return true
}

type Char struct {
	ObjectImpl
	Value rune
}

func (o *Char) String() string {
	return string(o.Value)
}

func (o *Char) TypeName() string {
	return "char"
}

func (o *Char) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Char:
		switch op {
		case token.Add:
			r := o.Value + rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Char{Value: r}, nil
		case token.Sub:
			r := o.Value - rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Char{Value: r}, nil
		case token.Less:
			if o.Value < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case *Int:
		switch op {
		case token.Add:
			r := o.Value + rune(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Char{Value: r}, nil
		case token.Sub:
			r := o.Value - rune(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Char{Value: r}, nil
		case token.Less:
			if int64(o.Value) < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if int64(o.Value) > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if int64(o.Value) <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if int64(o.Value) >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Char) Copy() Object {
	return &Char{Value: o.Value}
}

func (o *Char) IsFalsy() bool {
	return o.Value == 0
}

func (o *Char) Equals(x Object) bool {
	t, ok := x.(*Char)
	if !ok {
		return false
	}
	return o.Value == t.Value
}

type CompiledFunction struct {
	ObjectImpl
	Instructions  []byte
	NumLocals     int
	NumParameters int
	VarArgs       bool
	SourceMap     map[int]parser.Pos
	Free          []*ObjectPtr
}

func (o *CompiledFunction) TypeName() string {
	return "compiled-function"
}

func (o *CompiledFunction) String() string {
	return "<compiled-function>"
}

func (o *CompiledFunction) Copy() Object {
	return &CompiledFunction{
		Instructions:  append([]byte{}, o.Instructions...),
		NumLocals:     o.NumLocals,
		NumParameters: o.NumParameters,
		VarArgs:       o.VarArgs,
		Free:          append([]*ObjectPtr{}, o.Free...),
	}
}

func (o *CompiledFunction) Equals(_ Object) bool {
	return false
}

func (o *CompiledFunction) SourcePos(ip int) parser.Pos {
	for ip >= 0 {
		if p, ok := o.SourceMap[ip]; ok {
			return p
		}
		ip--
	}
	return parser.NoPos
}

func (o *CompiledFunction) CanCall() bool {
	return true
}

type Error struct {
	ObjectImpl
	Value Object
}

func (o *Error) TypeName() string {
	return "error"
}

func (o *Error) String() string {
	if o.Value != nil {
		return fmt.Sprintf("error: %s", o.Value.String())
	}
	return "error"
}

func (o *Error) IsFalsy() bool {
	return true
}

func (o *Error) Copy() Object {
	return &Error{Value: o.Value.Copy()}
}

func (o *Error) Equals(x Object) bool {
	return o == x
}

func (o *Error) IndexGet(index Object) (res Object, err error) {
	if strIdx, _ := ToString(index); strIdx != "value" {
		err = ErrInvalidIndexOnError
		return
	}
	res = o.Value
	return
}

type Float struct {
	ObjectImpl
	Value float64
}

func (o *Float) String() string {
	return strconv.FormatFloat(o.Value, 'f', -1, 64)
}

func (o *Float) TypeName() string {
	return "float"
}

func (o *Float) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Float:
		switch op {
		case token.Add:
			r := o.Value + rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Sub:
			r := o.Value - rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Mul:
			r := o.Value * rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Quo:
			r := o.Value / rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Less:
			if o.Value < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case *Int:
		switch op {
		case token.Add:
			r := o.Value + float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Sub:
			r := o.Value - float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Mul:
			r := o.Value * float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Quo:
			r := o.Value / float64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Float{Value: r}, nil
		case token.Less:
			if o.Value < float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= float64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Float) Copy() Object {
	return &Float{Value: o.Value}
}

func (o *Float) IsFalsy() bool {
	return math.IsNaN(o.Value)
}

func (o *Float) Equals(x Object) bool {
	t, ok := x.(*Float)
	if !ok {
		return false
	}
	return o.Value == t.Value
}

type ImmutableArray struct {
	ObjectImpl
	Value []Object
}

func (o *ImmutableArray) TypeName() string {
	return "immutable-array"
}

func (o *ImmutableArray) String() string {
	var elements []string
	for _, e := range o.Value {
		elements = append(elements, e.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (o *ImmutableArray) BinaryOp(op token.Token, rhs Object) (Object, error) {
	if rhs, ok := rhs.(*ImmutableArray); ok {
		switch op {
		case token.Add:
			return &Array{Value: append(o.Value, rhs.Value...)}, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *ImmutableArray) Copy() Object {
	var c []Object
	for _, elem := range o.Value {
		c = append(c, elem.Copy())
	}
	return &Array{Value: c}
}

func (o *ImmutableArray) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *ImmutableArray) Equals(x Object) bool {
	var xVal []Object
	switch x := x.(type) {
	case *Array:
		xVal = x.Value
	case *ImmutableArray:
		xVal = x.Value
	default:
		return false
	}
	if len(o.Value) != len(xVal) {
		return false
	}
	for i, e := range o.Value {
		if !e.Equals(xVal[i]) {
			return false
		}
	}
	return true
}

func (o *ImmutableArray) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	idxVal := int(intIdx.Value)
	if idxVal < 0 || idxVal >= len(o.Value) {
		res = UndefinedValue
		return
	}
	res = o.Value[idxVal]
	return
}

func (o *ImmutableArray) Iterate() Iterator {
	return &ArrayIterator{
		v: o.Value,
		l: len(o.Value),
	}
}

func (o *ImmutableArray) CanIterate() bool {
	return true
}

type ImmutableMap struct {
	ObjectImpl
	Value map[string]Object
}

func (o *ImmutableMap) TypeName() string {
	return "immutable-map"
}

func (o *ImmutableMap) String() string {
	var pairs []string
	for k, v := range o.Value {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (o *ImmutableMap) Copy() Object {
	c := make(map[string]Object)
	for k, v := range o.Value {
		c[k] = v.Copy()
	}
	return &Map{Value: c}
}

func (o *ImmutableMap) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *ImmutableMap) IndexGet(index Object) (res Object, err error) {
	strIdx, ok := ToString(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	res, ok = o.Value[strIdx]
	if !ok {
		res = UndefinedValue
	}
	return
}

func (o *ImmutableMap) Equals(x Object) bool {
	var xVal map[string]Object
	switch x := x.(type) {
	case *Map:
		xVal = x.Value
	case *ImmutableMap:
		xVal = x.Value
	default:
		return false
	}
	if len(o.Value) != len(xVal) {
		return false
	}
	for k, v := range o.Value {
		tv := xVal[k]
		if !v.Equals(tv) {
			return false
		}
	}
	return true
}

func (o *ImmutableMap) Iterate() Iterator {
	var keys []string
	for k := range o.Value {
		keys = append(keys, k)
	}
	return &MapIterator{
		v: o.Value,
		k: keys,
		l: len(keys),
	}
}

func (o *ImmutableMap) CanIterate() bool {
	return true
}

type Int struct {
	ObjectImpl
	Value int64
}

func (o *Int) String() string {
	return strconv.FormatInt(o.Value, 10)
}

func (o *Int) TypeName() string {
	return "int"
}

func (o *Int) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Int:
		switch op {
		case token.Add:
			r := o.Value + rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Sub:
			r := o.Value - rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Mul:
			r := o.Value * rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Quo:
			r := o.Value / rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Rem:
			r := o.Value % rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.And:
			r := o.Value & rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Or:
			r := o.Value | rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Xor:
			r := o.Value ^ rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.AndNot:
			r := o.Value &^ rhs.Value
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Shl:
			r := o.Value << uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Shr:
			r := o.Value >> uint64(rhs.Value)
			if r == o.Value {
				return o, nil
			}
			return &Int{Value: r}, nil
		case token.Less:
			if o.Value < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case *Float:
		switch op {
		case token.Add:
			return &Float{Value: float64(o.Value) + rhs.Value}, nil
		case token.Sub:
			return &Float{Value: float64(o.Value) - rhs.Value}, nil
		case token.Mul:
			return &Float{Value: float64(o.Value) * rhs.Value}, nil
		case token.Quo:
			return &Float{Value: float64(o.Value) / rhs.Value}, nil
		case token.Less:
			if float64(o.Value) < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if float64(o.Value) > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if float64(o.Value) <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if float64(o.Value) >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case *Char:
		switch op {
		case token.Add:
			return &Char{Value: rune(o.Value) + rhs.Value}, nil
		case token.Sub:
			return &Char{Value: rune(o.Value) - rhs.Value}, nil
		case token.Less:
			if o.Value < int64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value > int64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value <= int64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value >= int64(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Int) Copy() Object {
	return &Int{Value: o.Value}
}

func (o *Int) IsFalsy() bool {
	return o.Value == 0
}

func (o *Int) Equals(x Object) bool {
	t, ok := x.(*Int)
	if !ok {
		return false
	}
	return o.Value == t.Value
}

type Map struct {
	ObjectImpl
	Value map[string]Object
}

func (o *Map) TypeName() string {
	return "map"
}

func (o *Map) String() string {
	var pairs []string
	for k, v := range o.Value {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

func (o *Map) Copy() Object {
	c := make(map[string]Object)
	for k, v := range o.Value {
		c[k] = v.Copy()
	}
	return &Map{Value: c}
}

func (o *Map) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *Map) Equals(x Object) bool {
	var xVal map[string]Object
	switch x := x.(type) {
	case *Map:
		xVal = x.Value
	case *ImmutableMap:
		xVal = x.Value
	default:
		return false
	}
	if len(o.Value) != len(xVal) {
		return false
	}
	for k, v := range o.Value {
		tv := xVal[k]
		if !v.Equals(tv) {
			return false
		}
	}
	return true
}

func (o *Map) IndexGet(index Object) (res Object, err error) {
	strIdx, ok := ToString(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	res, ok = o.Value[strIdx]
	if !ok {
		res = UndefinedValue
	}
	return
}

func (o *Map) IndexSet(index, value Object) (err error) {
	strIdx, ok := ToString(index)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	o.Value[strIdx] = value
	return nil
}

func (o *Map) Iterate() Iterator {
	var keys []string
	for k := range o.Value {
		keys = append(keys, k)
	}
	return &MapIterator{
		v: o.Value,
		k: keys,
		l: len(keys),
	}
}

func (o *Map) CanIterate() bool {
	return true
}

type ObjectPtr struct {
	ObjectImpl
	Value *Object
}

func (o *ObjectPtr) String() string {
	return "free-var"
}

func (o *ObjectPtr) TypeName() string {
	return "<free-var>"
}

func (o *ObjectPtr) Copy() Object {
	return o
}

func (o *ObjectPtr) IsFalsy() bool {
	return o.Value == nil
}

func (o *ObjectPtr) Equals(x Object) bool {
	return o == x
}

type String struct {
	ObjectImpl
	Value   string
	runeStr []rune
}

func (o *String) TypeName() string {
	return "string"
}

func (o *String) String() string {
	return strconv.Quote(o.Value)
}

func (o *String) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch op {
	case token.Add:
		switch rhs := rhs.(type) {
		case *String:
			if len(o.Value)+len(rhs.Value) > MaxStringLen {
				return nil, ErrStringLimit
			}
			return &String{Value: o.Value + rhs.Value}, nil
		default:
			rhsStr := rhs.String()
			if len(o.Value)+len(rhsStr) > MaxStringLen {
				return nil, ErrStringLimit
			}
			return &String{Value: o.Value + rhsStr}, nil
		}
	case token.Less:
		switch rhs := rhs.(type) {
		case *String:
			if o.Value < rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case token.LessEq:
		switch rhs := rhs.(type) {
		case *String:
			if o.Value <= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case token.Greater:
		switch rhs := rhs.(type) {
		case *String:
			if o.Value > rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	case token.GreaterEq:
		switch rhs := rhs.(type) {
		case *String:
			if o.Value >= rhs.Value {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *String) IsFalsy() bool {
	return len(o.Value) == 0
}

func (o *String) Copy() Object {
	return &String{Value: o.Value}
}

func (o *String) Equals(x Object) bool {
	t, ok := x.(*String)
	if !ok {
		return false
	}
	return o.Value == t.Value
}

func (o *String) IndexGet(index Object) (res Object, err error) {
	intIdx, ok := index.(*Int)
	if !ok {
		err = ErrInvalidIndexType
		return
	}
	idxVal := int(intIdx.Value)
	if o.runeStr == nil {
		o.runeStr = []rune(o.Value)
	}
	if idxVal < 0 || idxVal >= len(o.runeStr) {
		res = UndefinedValue
		return
	}
	res = &Char{Value: o.runeStr[idxVal]}
	return
}

func (o *String) Iterate() Iterator {
	if o.runeStr == nil {
		o.runeStr = []rune(o.Value)
	}
	return &StringIterator{
		v: o.runeStr,
		l: len(o.runeStr),
	}
}

func (o *String) CanIterate() bool {
	return true
}

type Time struct {
	ObjectImpl
	Value time.Time
}

func (o *Time) String() string {
	return o.Value.String()
}

func (o *Time) TypeName() string {
	return "time"
}

func (o *Time) BinaryOp(op token.Token, rhs Object) (Object, error) {
	switch rhs := rhs.(type) {
	case *Int:
		switch op {
		case token.Add:
			if rhs.Value == 0 {
				return o, nil
			}
			return &Time{Value: o.Value.Add(time.Duration(rhs.Value))}, nil
		case token.Sub:
			if rhs.Value == 0 {
				return o, nil
			}
			return &Time{Value: o.Value.Add(time.Duration(-rhs.Value))}, nil
		}
	case *Time:
		switch op {
		case token.Sub:
			return &Int{Value: int64(o.Value.Sub(rhs.Value))}, nil
		case token.Less:
			if o.Value.Before(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.Greater:
			if o.Value.After(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.LessEq:
			if o.Value.Equal(rhs.Value) || o.Value.Before(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		case token.GreaterEq:
			if o.Value.Equal(rhs.Value) || o.Value.After(rhs.Value) {
				return TrueValue, nil
			}
			return FalseValue, nil
		}
	}
	return nil, ErrInvalidOperator
}

func (o *Time) Copy() Object {
	return &Time{Value: o.Value}
}

func (o *Time) IsFalsy() bool {
	return o.Value.IsZero()
}

func (o *Time) Equals(x Object) bool {
	t, ok := x.(*Time)
	if !ok {
		return false
	}
	return o.Value.Equal(t.Value)
}

type Undefined struct {
	ObjectImpl
}

func (o *Undefined) TypeName() string {
	return "undefined"
}

func (o *Undefined) String() string {
	return "<undefined>"
}

func (o *Undefined) Copy() Object {
	return o
}

func (o *Undefined) IsFalsy() bool {
	return true
}

func (o *Undefined) Equals(x Object) bool {
	return o == x
}

func (o *Undefined) IndexGet(_ Object) (Object, error) {
	return UndefinedValue, nil
}

func (o *Undefined) Iterate() Iterator {
	return o
}

func (o *Undefined) CanIterate() bool {
	return true
}

func (o *Undefined) Next() bool {
	return false
}

func (o *Undefined) Key() Object {
	return o
}

func (o *Undefined) Value() Object {
	return o
}

type UserFunction struct {
	ObjectImpl
	Name  string
	Value CallableFunc
}

func (o *UserFunction) TypeName() string {
	return "user-function:" + o.Name
}

func (o *UserFunction) String() string {
	return "<user-function>"
}

func (o *UserFunction) Copy() Object {
	return &UserFunction{Value: o.Value, Name: o.Name}
}

func (o *UserFunction) Equals(_ Object) bool {
	return false
}

func (o *UserFunction) Call(args ...Object) (Object, error) {
	return o.Value(args...)
}

func (o *UserFunction) CanCall() bool {
	return true
}
