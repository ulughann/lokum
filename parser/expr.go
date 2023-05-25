package parser

import (
	"strings"

	"github.com/onrirr/lokum/token"
)

type Expr interface {
	Node
	exprNode()
}

type ArrayLit struct {
	Elements []Expr
	LBrack   Pos
	RBrack   Pos
}

func (e *ArrayLit) exprNode() {}

func (e *ArrayLit) Pos() Pos {
	return e.LBrack
}

func (e *ArrayLit) End() Pos {
	return e.RBrack + 1
}

func (e *ArrayLit) String() string {
	var elements []string
	for _, m := range e.Elements {
		elements = append(elements, m.String())
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

type BadExpr struct {
	From Pos
	To   Pos
}

func (e *BadExpr) exprNode() {}

func (e *BadExpr) Pos() Pos {
	return e.From
}

func (e *BadExpr) End() Pos {
	return e.To
}

func (e *BadExpr) String() string {
	return "<kötü ifade>"
}

type BinaryExpr struct {
	LHS      Expr
	RHS      Expr
	Token    token.Token
	TokenPos Pos
}

func (e *BinaryExpr) exprNode() {}

func (e *BinaryExpr) Pos() Pos {
	return e.LHS.Pos()
}

func (e *BinaryExpr) End() Pos {
	return e.RHS.End()
}

func (e *BinaryExpr) String() string {
	return "(" + e.LHS.String() + " " + e.Token.String() +
		" " + e.RHS.String() + ")"
}

type BoolLit struct {
	Value    bool
	ValuePos Pos
	Literal  string
}

func (e *BoolLit) exprNode() {}

func (e *BoolLit) Pos() Pos {
	return e.ValuePos
}

func (e *BoolLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *BoolLit) String() string {
	return e.Literal
}

type CallExpr struct {
	Func     Expr
	LParen   Pos
	Args     []Expr
	Ellipsis Pos
	RParen   Pos
}

func (e *CallExpr) exprNode() {}

func (e *CallExpr) Pos() Pos {
	return e.Func.Pos()
}

func (e *CallExpr) End() Pos {
	return e.RParen + 1
}

func (e *CallExpr) String() string {
	var args []string
	for _, e := range e.Args {
		args = append(args, e.String())
	}
	if len(args) > 0 && e.Ellipsis.IsValid() {
		args[len(args)-1] = args[len(args)-1] + "..."
	}
	return e.Func.String() + "(" + strings.Join(args, ", ") + ")"
}

type CharLit struct {
	Value    rune
	ValuePos Pos
	Literal  string
}

func (e *CharLit) exprNode() {}

func (e *CharLit) Pos() Pos {
	return e.ValuePos
}

func (e *CharLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *CharLit) String() string {
	return e.Literal
}

type CondExpr struct {
	Cond        Expr
	True        Expr
	False       Expr
	QuestionPos Pos
	ColonPos    Pos
}

func (e *CondExpr) exprNode() {}

func (e *CondExpr) Pos() Pos {
	return e.Cond.Pos()
}

func (e *CondExpr) End() Pos {
	return e.False.End()
}

func (e *CondExpr) String() string {
	return "(" + e.Cond.String() + " ? " + e.True.String() +
		" : " + e.False.String() + ")"
}

type ErrorExpr struct {
	Expr     Expr
	ErrorPos Pos
	LParen   Pos
	RParen   Pos
}

func (e *ErrorExpr) exprNode() {}

func (e *ErrorExpr) Pos() Pos {
	return e.ErrorPos
}

func (e *ErrorExpr) End() Pos {
	return e.RParen
}

func (e *ErrorExpr) String() string {
	return "hata(" + e.Expr.String() + ")"
}

type FloatLit struct {
	Value    float64
	ValuePos Pos
	Literal  string
}

func (e *FloatLit) exprNode() {}

func (e *FloatLit) Pos() Pos {
	return e.ValuePos
}

func (e *FloatLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *FloatLit) String() string {
	return e.Literal
}

type FuncLit struct {
	Type *FuncType
	Body *BlockStmt
}

func (e *FuncLit) exprNode() {}

func (e *FuncLit) Pos() Pos {
	return e.Type.Pos()
}

func (e *FuncLit) End() Pos {
	return e.Body.End()
}

func (e *FuncLit) String() string {
	return "fn" + e.Type.Params.String() + " " + e.Body.String()
}

type FuncType struct {
	FuncPos Pos
	Params  *IdentList
}

func (e *FuncType) exprNode() {}

func (e *FuncType) Pos() Pos {
	return e.FuncPos
}

func (e *FuncType) End() Pos {
	return e.Params.End()
}

func (e *FuncType) String() string {
	return "fn" + e.Params.String()
}

type Ident struct {
	Name    string
	NamePos Pos
}

func (e *Ident) exprNode() {}

func (e *Ident) Pos() Pos {
	return e.NamePos
}

func (e *Ident) End() Pos {
	return Pos(int(e.NamePos) + len(e.Name))
}

func (e *Ident) String() string {
	if e != nil {
		return e.Name
	}
	return nullRep
}

type ImmutableExpr struct {
	Expr     Expr
	ErrorPos Pos
	LParen   Pos
	RParen   Pos
}

func (e *ImmutableExpr) exprNode() {}

func (e *ImmutableExpr) Pos() Pos {
	return e.ErrorPos
}

func (e *ImmutableExpr) End() Pos {
	return e.RParen
}

func (e *ImmutableExpr) String() string {
	return "sabit(" + e.Expr.String() + ")"
}

type ImportExpr struct {
	ModuleName string
	Token      token.Token
	TokenPos   Pos
}

func (e *ImportExpr) exprNode() {}

func (e *ImportExpr) Pos() Pos {
	return e.TokenPos
}

func (e *ImportExpr) End() Pos {

	return Pos(int(e.TokenPos) + 10 + len(e.ModuleName))
}

func (e *ImportExpr) String() string {
	return `kullan("` + e.ModuleName + `")"`
}

type IndexExpr struct {
	Expr   Expr
	LBrack Pos
	Index  Expr
	RBrack Pos
}

func (e *IndexExpr) exprNode() {}

func (e *IndexExpr) Pos() Pos {
	return e.Expr.Pos()
}

func (e *IndexExpr) End() Pos {
	return e.RBrack + 1
}

func (e *IndexExpr) String() string {
	var index string
	if e.Index != nil {
		index = e.Index.String()
	}
	return e.Expr.String() + "[" + index + "]"
}

type IntLit struct {
	Value    int64
	ValuePos Pos
	Literal  string
}

func (e *IntLit) exprNode() {}

func (e *IntLit) Pos() Pos {
	return e.ValuePos
}

func (e *IntLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *IntLit) String() string {
	return e.Literal
}

type MapElementLit struct {
	Key      string
	KeyPos   Pos
	ColonPos Pos
	Value    Expr
}

func (e *MapElementLit) exprNode() {}

func (e *MapElementLit) Pos() Pos {
	return e.KeyPos
}

func (e *MapElementLit) End() Pos {
	return e.Value.End()
}

func (e *MapElementLit) String() string {
	return e.Key + ": " + e.Value.String()
}

type MapLit struct {
	LBrace   Pos
	Elements []*MapElementLit
	RBrace   Pos
}

func (e *MapLit) exprNode() {}

func (e *MapLit) Pos() Pos {
	return e.LBrace
}

func (e *MapLit) End() Pos {
	return e.RBrace + 1
}

func (e *MapLit) String() string {
	var elements []string
	for _, m := range e.Elements {
		elements = append(elements, m.String())
	}
	return "{" + strings.Join(elements, ", ") + "}"
}

type ParenExpr struct {
	Expr   Expr
	LParen Pos
	RParen Pos
}

func (e *ParenExpr) exprNode() {}

func (e *ParenExpr) Pos() Pos {
	return e.LParen
}

func (e *ParenExpr) End() Pos {
	return e.RParen + 1
}

func (e *ParenExpr) String() string {
	return "(" + e.Expr.String() + ")"
}

type SelectorExpr struct {
	Expr Expr
	Sel  Expr
}

func (e *SelectorExpr) exprNode() {}

func (e *SelectorExpr) Pos() Pos {
	return e.Expr.Pos()
}

func (e *SelectorExpr) End() Pos {
	return e.Sel.End()
}

func (e *SelectorExpr) String() string {
	return e.Expr.String() + "." + e.Sel.String()
}

type SliceExpr struct {
	Expr   Expr
	LBrack Pos
	Low    Expr
	High   Expr
	RBrack Pos
}

func (e *SliceExpr) exprNode() {}

func (e *SliceExpr) Pos() Pos {
	return e.Expr.Pos()
}

func (e *SliceExpr) End() Pos {
	return e.RBrack + 1
}

func (e *SliceExpr) String() string {
	var low, high string
	if e.Low != nil {
		low = e.Low.String()
	}
	if e.High != nil {
		high = e.High.String()
	}
	return e.Expr.String() + "[" + low + ":" + high + "]"
}

type StringLit struct {
	Value    string
	ValuePos Pos
	Literal  string
}

func (e *StringLit) exprNode() {}

func (e *StringLit) Pos() Pos {
	return e.ValuePos
}

func (e *StringLit) End() Pos {
	return Pos(int(e.ValuePos) + len(e.Literal))
}

func (e *StringLit) String() string {
	return e.Literal
}

type UnaryExpr struct {
	Expr     Expr
	Token    token.Token
	TokenPos Pos
}

func (e *UnaryExpr) exprNode() {}

func (e *UnaryExpr) Pos() Pos {
	return e.Expr.Pos()
}

func (e *UnaryExpr) End() Pos {
	return e.Expr.End()
}

func (e *UnaryExpr) String() string {
	return "(" + e.Token.String() + e.Expr.String() + ")"
}

type UndefinedLit struct {
	TokenPos Pos
}

func (e *UndefinedLit) exprNode() {}

func (e *UndefinedLit) Pos() Pos {
	return e.TokenPos
}

func (e *UndefinedLit) End() Pos {
	return e.TokenPos + 9
}

func (e *UndefinedLit) String() string {
	return "tanımsız"
}
