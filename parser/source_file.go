package parser

import (
	"fmt"
	"sort"
)

type SourceFilePos struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

func (p SourceFilePos) IsValid() bool {
	return p.Line > 0
}

func (p SourceFilePos) String() string {
	s := p.Filename
	if p.IsValid() {
		if s != "" {
			s += ":"
		}
		s += fmt.Sprintf("%d", p.Line)
		if p.Column != 0 {
			s += fmt.Sprintf(":%d", p.Column)
		}
	}
	if s == "" {
		s = "-"
	}
	return s
}

type SourceFileSet struct {
	Base     int
	Files    []*SourceFile
	LastFile *SourceFile
}

func NewFileSet() *SourceFileSet {
	return &SourceFileSet{
		Base: 1,
	}
}

func (s *SourceFileSet) AddFile(filename string, base, size int) *SourceFile {
	if base < 0 {
		base = s.Base
	}
	if base < s.Base || size < 0 {
		panic("geçersiz dosya boyutu")
	}
	f := &SourceFile{
		set:   s,
		Name:  filename,
		Base:  base,
		Size:  size,
		Lines: []int{0},
	}
	base += size + 1
	if base < 0 {
		panic("overflow")
	}

	s.Base = base
	s.Files = append(s.Files, f)
	s.LastFile = f
	return f
}

func (s *SourceFileSet) File(p Pos) (f *SourceFile) {
	if p != NoPos {
		f = s.file(p)
	}
	return
}

func (s *SourceFileSet) Position(p Pos) (pos SourceFilePos) {
	if p != NoPos {
		if f := s.file(p); f != nil {
			return f.position(p)
		}
	}
	return
}

func (s *SourceFileSet) file(p Pos) *SourceFile {

	f := s.LastFile
	if f != nil && f.Base <= int(p) && int(p) <= f.Base+f.Size {
		return f
	}

	if i := searchFiles(s.Files, int(p)); i >= 0 {
		f := s.Files[i]

		if int(p) <= f.Base+f.Size {
			s.LastFile = f
			return f
		}
	}
	return nil
}

func searchFiles(a []*SourceFile, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i].Base > x }) - 1
}

type SourceFile struct {
	set *SourceFileSet

	Name string

	Base int

	Size int

	Lines []int
}

func (f *SourceFile) Set() *SourceFileSet {
	return f.set
}

func (f *SourceFile) LineCount() int {
	return len(f.Lines)
}

func (f *SourceFile) AddLine(offset int) {
	i := len(f.Lines)
	if (i == 0 || f.Lines[i-1] < offset) && offset < f.Size {
		f.Lines = append(f.Lines, offset)
	}
}

func (f *SourceFile) LineStart(line int) Pos {
	if line < 1 {
		panic("geçersiz satır sayısı (satır sayısı başlangıçı: 1)")
	}
	if line > len(f.Lines) {
		panic("geçersiz satır sayısı")
	}
	return Pos(f.Base + f.Lines[line-1])
}

func (f *SourceFile) FileSetPos(offset int) Pos {
	if offset > f.Size {
		panic("illegal")
	}
	return Pos(f.Base + offset)
}

func (f *SourceFile) Offset(p Pos) int {
	if int(p) < f.Base || int(p) > f.Base+f.Size {
		panic("illegal")
	}
	return int(p) - f.Base
}

func (f *SourceFile) Position(p Pos) (pos SourceFilePos) {
	if p != NoPos {
		if int(p) < f.Base || int(p) > f.Base+f.Size {
			panic("illegal")
		}
		pos = f.position(p)
	}
	return
}

func (f *SourceFile) position(p Pos) (pos SourceFilePos) {
	offset := int(p) - f.Base
	pos.Offset = offset
	pos.Filename, pos.Line, pos.Column = f.unpack(offset)
	return
}

func (f *SourceFile) unpack(offset int) (filename string, line, column int) {
	filename = f.Name
	if i := searchInts(f.Lines, offset); i >= 0 {
		line, column = i+1, offset-f.Lines[i]+1
	}
	return
}

func searchInts(a []int, x int) int {

	i, j := 0, len(a)
	for i < j {
		h := i + (j-i)/2

		if a[h] <= x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i - 1
}
