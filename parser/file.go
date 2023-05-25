package parser

import (
	"strings"
)

type File struct {
	InputFile *SourceFile
	Stmts     []Stmt
}

func (n *File) Pos() Pos {
	return Pos(n.InputFile.Base)
}

func (n *File) End() Pos {
	return Pos(n.InputFile.Base + n.InputFile.Size)
}

func (n *File) String() string {
	var stmts []string
	for _, e := range n.Stmts {
		stmts = append(stmts, e.String())
	}
	return strings.Join(stmts, "; ")
}
