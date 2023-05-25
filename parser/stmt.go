package parser

import (
	"strings"

	"github.com/onrirr/lokum/token"
)

type Stmt interface {
	Node
	stmtNode()
}

type AssignStmt struct {
	LHS      []Expr
	RHS      []Expr
	Token    token.Token
	TokenPos Pos
}

func (s *AssignStmt) stmtNode() {}

func (s *AssignStmt) Pos() Pos {
	return s.LHS[0].Pos()
}

func (s *AssignStmt) End() Pos {
	return s.RHS[len(s.RHS)-1].End()
}

func (s *AssignStmt) String() string {
	var lhs, rhs []string
	for _, e := range s.LHS {
		lhs = append(lhs, e.String())
	}
	for _, e := range s.RHS {
		rhs = append(rhs, e.String())
	}
	return strings.Join(lhs, ", ") + " " + s.Token.String() +
		" " + strings.Join(rhs, ", ")
}

type BadStmt struct {
	From Pos
	To   Pos
}

func (s *BadStmt) stmtNode() {}

func (s *BadStmt) Pos() Pos {
	return s.From
}

func (s *BadStmt) End() Pos {
	return s.To
}

func (s *BadStmt) String() string {
	return "<kötü ifade>"
}

type BlockStmt struct {
	Stmts  []Stmt
	LBrace Pos
	RBrace Pos
}

func (s *BlockStmt) stmtNode() {}

func (s *BlockStmt) Pos() Pos {
	return s.LBrace
}

func (s *BlockStmt) End() Pos {
	return s.RBrace + 1
}

func (s *BlockStmt) String() string {
	var list []string
	for _, e := range s.Stmts {
		list = append(list, e.String())
	}
	return "{" + strings.Join(list, "; ") + "}"
}

type BranchStmt struct {
	Token    token.Token
	TokenPos Pos
	Label    *Ident
}

func (s *BranchStmt) stmtNode() {}

func (s *BranchStmt) Pos() Pos {
	return s.TokenPos
}

func (s *BranchStmt) End() Pos {
	if s.Label != nil {
		return s.Label.End()
	}

	return Pos(int(s.TokenPos) + len(s.Token.String()))
}

func (s *BranchStmt) String() string {
	var label string
	if s.Label != nil {
		label = " " + s.Label.Name
	}
	return s.Token.String() + label
}

type EmptyStmt struct {
	Semicolon Pos
	Implicit  bool
}

func (s *EmptyStmt) stmtNode() {}

func (s *EmptyStmt) Pos() Pos {
	return s.Semicolon
}

func (s *EmptyStmt) End() Pos {
	if s.Implicit {
		return s.Semicolon
	}
	return s.Semicolon + 1
}

func (s *EmptyStmt) String() string {
	return ";"
}

type ExportStmt struct {
	ExportPos Pos
	Result    Expr
}

func (s *ExportStmt) stmtNode() {}

func (s *ExportStmt) Pos() Pos {
	return s.ExportPos
}

func (s *ExportStmt) End() Pos {
	return s.Result.End()
}

func (s *ExportStmt) String() string {
	return "paylaş " + s.Result.String()
}

type ExprStmt struct {
	Expr Expr
}

func (s *ExprStmt) stmtNode() {}

func (s *ExprStmt) Pos() Pos {
	return s.Expr.Pos()
}

func (s *ExprStmt) End() Pos {
	return s.Expr.End()
}

func (s *ExprStmt) String() string {
	return s.Expr.String()
}

type ForInStmt struct {
	ForPos   Pos
	Key      *Ident
	Value    *Ident
	Iterable Expr
	Body     *BlockStmt
}

func (s *ForInStmt) stmtNode() {}

func (s *ForInStmt) Pos() Pos {
	return s.ForPos
}

func (s *ForInStmt) End() Pos {
	return s.Body.End()
}

func (s *ForInStmt) String() string {
	if s.Value != nil {
		return "tekrarla " + s.Key.String() + ", " + s.Value.String() +
			" in " + s.Iterable.String() + " " + s.Body.String()
	}
	return "tekrarla " + s.Key.String() + " in " + s.Iterable.String() +
		" " + s.Body.String()
}

type ForStmt struct {
	ForPos Pos
	Init   Stmt
	Cond   Expr
	Post   Stmt
	Body   *BlockStmt
}

func (s *ForStmt) stmtNode() {}

func (s *ForStmt) Pos() Pos {
	return s.ForPos
}

func (s *ForStmt) End() Pos {
	return s.Body.End()
}

func (s *ForStmt) String() string {
	var init, cond, post string
	if s.Init != nil {
		init = s.Init.String()
	}
	if s.Cond != nil {
		cond = s.Cond.String() + " "
	}
	if s.Post != nil {
		post = s.Post.String()
	}

	if init != "" || post != "" {
		return "tekrarla " + init + " ; " + cond + " ; " + post + s.Body.String()
	}
	return "tekrarla " + cond + s.Body.String()
}

type IfStmt struct {
	IfPos Pos
	Init  Stmt
	Cond  Expr
	Body  *BlockStmt
	Else  Stmt
}

func (s *IfStmt) stmtNode() {}

func (s *IfStmt) Pos() Pos {
	return s.IfPos
}

func (s *IfStmt) End() Pos {
	if s.Else != nil {
		return s.Else.End()
	}
	return s.Body.End()
}

func (s *IfStmt) String() string {
	var initStmt, elseStmt string
	if s.Init != nil {
		initStmt = s.Init.String() + "; "
	}
	if s.Else != nil {
		elseStmt = " yoksa " + s.Else.String()
	}
	return "eğer " + initStmt + s.Cond.String() + " " +
		s.Body.String() + elseStmt
}

type IncDecStmt struct {
	Expr     Expr
	Token    token.Token
	TokenPos Pos
}

func (s *IncDecStmt) stmtNode() {}

func (s *IncDecStmt) Pos() Pos {
	return s.Expr.Pos()
}

func (s *IncDecStmt) End() Pos {
	return Pos(int(s.TokenPos) + 2)
}

func (s *IncDecStmt) String() string {
	return s.Expr.String() + s.Token.String()
}

type ReturnStmt struct {
	ReturnPos Pos
	Result    Expr
}

func (s *ReturnStmt) stmtNode() {}

func (s *ReturnStmt) Pos() Pos {
	return s.ReturnPos
}

func (s *ReturnStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.ReturnPos + 6
}

func (s *ReturnStmt) String() string {
	if s.Result != nil {
		return "dön " + s.Result.String()
	}
	return "dön"
}
