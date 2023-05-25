package token

import "strconv"

var keywords map[string]Token

type Token int

const (
	Illegal Token = iota
	EOF
	Comment
	_literalBeg
	Ident
	Int
	Float
	Char
	String
	_literalEnd
	_operatorBeg
	Add
	Sub
	Mul
	Quo
	Rem
	And
	Or
	Xor
	Shl
	Shr
	AndNot
	AddAssign
	SubAssign
	MulAssign
	QuoAssign
	RemAssign
	AndAssign
	OrAssign
	XorAssign
	ShlAssign
	ShrAssign
	AndNotAssign
	LAnd
	LOr
	Inc
	Dec
	Equal
	Less
	Greater
	Assign
	Not
	NotEqual
	LessEq
	GreaterEq
	Define
	Ellipsis
	LParen
	LBrack
	LBrace
	Comma
	Period
	RParen
	RBrack
	RBrace
	Semicolon
	Colon
	Question
	_operatorEnd
	_keywordBeg
	Break
	Continue
	Else
	For
	Func
	Error
	Immutable
	If
	Return
	Export
	True
	False
	In
	Undefined
	Import
	_keywordEnd
)

var tokens = [...]string{
	Illegal:      "YASAKLI",
	EOF:          "DOSYA_SONU",
	Comment:      "YORUM",
	Ident:        "BOŞLUK",
	Int:          "SAYI",
	Float:        "FLOAT",
	Char:         "KARAKTER",
	String:       "YAZI",
	Add:          "+",
	Sub:          "-",
	Mul:          "*",
	Quo:          "/",
	Rem:          "%",
	And:          "&",
	Or:           "|",
	Xor:          "^",
	Shl:          "<<",
	Shr:          ">>",
	AndNot:       "&^",
	AddAssign:    "+=",
	SubAssign:    "-=",
	MulAssign:    "*=",
	QuoAssign:    "/=",
	RemAssign:    "%=",
	AndAssign:    "&=",
	OrAssign:     "|=",
	XorAssign:    "^=",
	ShlAssign:    "<<=",
	ShrAssign:    ">>=",
	AndNotAssign: "&^=",
	LAnd:         "&&",
	LOr:          "||",
	Inc:          "++",
	Dec:          "--",
	Equal:        "==",
	Less:         "<",
	Greater:      ">",
	Assign:       "=",
	Not:          "!",
	NotEqual:     "!=",
	LessEq:       "<=",
	GreaterEq:    ">=",
	Define:       ":=",
	Ellipsis:     "...",
	LParen:       "(",
	LBrack:       "[",
	LBrace:       "{",
	Comma:        ",",
	Period:       ".",
	RParen:       ")",
	RBrack:       "]",
	RBrace:       "}",
	Semicolon:    ";",
	Colon:        ":",
	Question:     "?",
	Break:        "dur",
	Continue:     "devam",
	Else:         "yoksa",
	For:          "tekrarla",
	Func:         "fn",
	Error:        "hata",
	Immutable:    "sabit",
	If:           "eğer",
	Return:       "dön",
	Export:       "paylaş",
	True:         "doğru",
	False:        "yanlış",
	In:           "in",
	Undefined:    "tanımsız",
	Import:       "kullan",
}

func (tok Token) String() string {
	s := ""

	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}

	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}

	return s
}

const LowestPrec = 0

func (tok Token) Precedence() int {
	switch tok {
	case LOr:
		return 1
	case LAnd:
		return 2
	case Equal, NotEqual, Less, LessEq, Greater, GreaterEq:
		return 3
	case Add, Sub, Or, Xor:
		return 4
	case Mul, Quo, Rem, Shl, Shr, And, AndNot:
		return 5
	}
	return LowestPrec
}

func (tok Token) IsLiteral() bool {
	return _literalBeg < tok && tok < _literalEnd
}

func (tok Token) IsOperator() bool {
	return _operatorBeg < tok && tok < _operatorEnd
}

func (tok Token) IsKeyword() bool {
	return _keywordBeg < tok && tok < _keywordEnd
}

func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return Ident
}

func init() {
	keywords = make(map[string]Token)
	for i := _keywordBeg + 1; i < _keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
}
