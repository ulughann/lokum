package lokum

type Iterator interface {
	Object

	Next() bool

	Key() Object

	Value() Object
}

type ArrayIterator struct {
	ObjectImpl
	v []Object
	i int
	l int
}

func (i *ArrayIterator) TypeName() string {
	return "array-iterator"
}

func (i *ArrayIterator) String() string {
	return "<array-iterator>"
}

func (i *ArrayIterator) IsFalsy() bool {
	return true
}

func (i *ArrayIterator) Equals(Object) bool {
	return false
}

func (i *ArrayIterator) Copy() Object {
	return &ArrayIterator{v: i.v, i: i.i, l: i.l}
}

func (i *ArrayIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

func (i *ArrayIterator) Key() Object {
	return &Int{Value: int64(i.i - 1)}
}

func (i *ArrayIterator) Value() Object {
	return i.v[i.i-1]
}

type BytesIterator struct {
	ObjectImpl
	v []byte
	i int
	l int
}

func (i *BytesIterator) TypeName() string {
	return "bytes-iterator"
}

func (i *BytesIterator) String() string {
	return "<bytes-iterator>"
}

func (i *BytesIterator) Equals(Object) bool {
	return false
}

func (i *BytesIterator) Copy() Object {
	return &BytesIterator{v: i.v, i: i.i, l: i.l}
}

func (i *BytesIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

func (i *BytesIterator) Key() Object {
	return &Int{Value: int64(i.i - 1)}
}

func (i *BytesIterator) Value() Object {
	return &Int{Value: int64(i.v[i.i-1])}
}

type MapIterator struct {
	ObjectImpl
	v map[string]Object
	k []string
	i int
	l int
}

func (i *MapIterator) TypeName() string {
	return "map-iterator"
}

func (i *MapIterator) String() string {
	return "<map-iterator>"
}

func (i *MapIterator) IsFalsy() bool {
	return true
}

func (i *MapIterator) Equals(Object) bool {
	return false
}

func (i *MapIterator) Copy() Object {
	return &MapIterator{v: i.v, k: i.k, i: i.i, l: i.l}
}

func (i *MapIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

func (i *MapIterator) Key() Object {
	k := i.k[i.i-1]
	return &String{Value: k}
}

func (i *MapIterator) Value() Object {
	k := i.k[i.i-1]
	return i.v[k]
}

type StringIterator struct {
	ObjectImpl
	v []rune
	i int
	l int
}

func (i *StringIterator) TypeName() string {
	return "string-iterator"
}

func (i *StringIterator) String() string {
	return "<string-iterator>"
}

func (i *StringIterator) IsFalsy() bool {
	return true
}

func (i *StringIterator) Equals(Object) bool {
	return false
}

func (i *StringIterator) Copy() Object {
	return &StringIterator{v: i.v, i: i.i, l: i.l}
}

func (i *StringIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

func (i *StringIterator) Key() Object {
	return &Int{Value: int64(i.i - 1)}
}

func (i *StringIterator) Value() Object {
	return &Char{Value: i.v[i.i-1]}
}
