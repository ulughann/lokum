package lokum

import (
	"errors"
)

type Variable struct {
	name  string
	value Object
}

func NewVariable(name string, value interface{}) (*Variable, error) {
	obj, err := FromInterface(value)
	if err != nil {
		return nil, err
	}
	return &Variable{
		name:  name,
		value: obj,
	}, nil
}

func (v *Variable) Name() string {
	return v.name
}

func (v *Variable) Value() interface{} {
	return ToInterface(v.value)
}

func (v *Variable) ValueType() string {
	return v.value.TypeName()
}

func (v *Variable) Int() int {
	c, _ := ToInt(v.value)
	return c
}

func (v *Variable) Int64() int64 {
	c, _ := ToInt64(v.value)
	return c
}

func (v *Variable) Float() float64 {
	c, _ := ToFloat64(v.value)
	return c
}

func (v *Variable) Char() rune {
	c, _ := ToRune(v.value)
	return c
}

func (v *Variable) Bool() bool {
	c, _ := ToBool(v.value)
	return c
}

func (v *Variable) Array() []interface{} {
	switch val := v.value.(type) {
	case *Array:
		var arr []interface{}
		for _, e := range val.Value {
			arr = append(arr, ToInterface(e))
		}
		return arr
	}
	return nil
}

func (v *Variable) Map() map[string]interface{} {
	switch val := v.value.(type) {
	case *Map:
		kv := make(map[string]interface{})
		for mk, mv := range val.Value {
			kv[mk] = ToInterface(mv)
		}
		return kv
	}
	return nil
}

func (v *Variable) String() string {
	c, _ := ToString(v.value)
	return c
}

func (v *Variable) Bytes() []byte {
	c, _ := ToByteSlice(v.value)
	return c
}

func (v *Variable) Error() error {
	err, ok := v.value.(*Error)
	if ok {
		return errors.New(err.String())
	}
	return nil
}

func (v *Variable) Object() Object {
	return v.value
}

func (v *Variable) IsUndefined() bool {
	return v.value == UndefinedValue
}
