package parser

import (
	"strings"
)

const (
	nullRep = "<null>"
)

type Node interface {
	Pos() Pos
	End() Pos
	String() string
}

type IdentList struct {
	LParen  Pos
	VarArgs bool
	List    []*Ident
	RParen  Pos
}

func (n *IdentList) Pos() Pos {
	if n.LParen.IsValid() {
		return n.LParen
	}
	if len(n.List) > 0 {
		return n.List[0].Pos()
	}
	return NoPos
}

func (n *IdentList) End() Pos {
	if n.RParen.IsValid() {
		return n.RParen + 1
	}
	if l := len(n.List); l > 0 {
		return n.List[l-1].End()
	}
	return NoPos
}

func (n *IdentList) NumFields() int {
	if n == nil {
		return 0
	}
	return len(n.List)
}

func (n *IdentList) String() string {
	var list []string
	for i, e := range n.List {
		if n.VarArgs && i == len(n.List)-1 {
			list = append(list, "..."+e.String())
		} else {
			list = append(list, e.String())
		}
	}
	return "(" + strings.Join(list, ", ") + ")"
}
