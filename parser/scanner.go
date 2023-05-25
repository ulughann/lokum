package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/onrirr/lokum/token"
)

const bom = 0xFEFF

type ScanMode int

const (
	ScanComments ScanMode = 1 << iota
	DontInsertSemis
)

type ScannerErrorHandler func(pos SourceFilePos, msg string)

type Scanner struct {
	file         *SourceFile
	src          []byte
	ch           rune
	offset       int
	readOffset   int
	lineOffset   int
	insertSemi   bool
	errorHandler ScannerErrorHandler
	errorCount   int
	mode         ScanMode
}

func NewScanner(
	file *SourceFile,
	src []byte,
	errorHandler ScannerErrorHandler,
	mode ScanMode,
) *Scanner {
	if file.Size != len(src) {
		panic(fmt.Sprintf("(%d) dosya boyutu src len ile uyuşmuyor (%d)",
			file.Size, len(src)))
	}

	s := &Scanner{
		file:         file,
		src:          src,
		errorHandler: errorHandler,
		ch:           ' ',
		mode:         mode,
	}

	s.next()
	if s.ch == bom {
		s.next()
	}

	return s
}

func (s *Scanner) ErrorCount() int {
	return s.errorCount
}

func (s *Scanner) Scan() (
	tok token.Token,
	literal string,
	pos Pos,
) {
	s.skipWhitespace()

	pos = s.file.FileSetPos(s.offset)

	insertSemi := false

	switch ch := s.ch; {
	case isLetter(ch):
		literal = s.scanIdentifier()
		tok = token.Lookup(literal)
		switch tok {
		case token.Ident, token.Break, token.Continue, token.Return,
			token.Export, token.True, token.False, token.Undefined:
			insertSemi = true
		}
	case ('0' <= ch && ch <= '9') || (ch == '.' && '0' <= s.peek() && s.peek() <= '9'):
		insertSemi = true
		tok, literal = s.scanNumber()
	default:
		s.next()

		switch ch {
		case -1:
			if s.insertSemi {
				s.insertSemi = false
				return token.Semicolon, "\n", pos
			}
			tok = token.EOF
		case '\n':

			s.insertSemi = false
			return token.Semicolon, "\n", pos
		case '"':
			insertSemi = true
			tok = token.String
			literal = s.scanString()
		case '\'':
			insertSemi = true
			tok = token.Char
			literal = s.scanRune()
		case '`':
			insertSemi = true
			tok = token.String
			literal = s.scanRawString()
		case ':':
			tok = s.switch2(token.Colon, token.Define)
		case '.':
			tok = token.Period
			if s.ch == '.' && s.peek() == '.' {
				s.next()
				s.next()
				tok = token.Ellipsis
			}
		case ',':
			tok = token.Comma
		case '?':
			tok = token.Question
		case ';':
			tok = token.Semicolon
			literal = ";"
		case '(':
			tok = token.LParen
		case ')':
			insertSemi = true
			tok = token.RParen
		case '[':
			tok = token.LBrack
		case ']':
			insertSemi = true
			tok = token.RBrack
		case '{':
			tok = token.LBrace
		case '}':
			insertSemi = true
			tok = token.RBrace
		case '+':
			tok = s.switch3(token.Add, token.AddAssign, '+', token.Inc)
			if tok == token.Inc {
				insertSemi = true
			}
		case '-':
			tok = s.switch3(token.Sub, token.SubAssign, '-', token.Dec)
			if tok == token.Dec {
				insertSemi = true
			}
		case '*':
			tok = s.switch2(token.Mul, token.MulAssign)
		case '/':
			if s.ch == '/' || s.ch == '*' {

				if s.insertSemi && s.findLineEnd() {

					s.ch = '/'
					s.offset = s.file.Offset(pos)
					s.readOffset = s.offset + 1
					s.insertSemi = false
					return token.Semicolon, "\n", pos
				}
				comment := s.scanComment()
				if s.mode&ScanComments == 0 {

					s.insertSemi = false
					return s.Scan()
				}
				tok = token.Comment
				literal = comment
			} else {
				tok = s.switch2(token.Quo, token.QuoAssign)
			}
		case '%':
			tok = s.switch2(token.Rem, token.RemAssign)
		case '^':
			tok = s.switch2(token.Xor, token.XorAssign)
		case '<':
			tok = s.switch4(token.Less, token.LessEq, '<',
				token.Shl, token.ShlAssign)
		case '>':
			tok = s.switch4(token.Greater, token.GreaterEq, '>',
				token.Shr, token.ShrAssign)
		case '=':
			tok = s.switch2(token.Assign, token.Equal)
		case '!':
			tok = s.switch2(token.Not, token.NotEqual)
		case '&':
			if s.ch == '^' {
				s.next()
				tok = s.switch2(token.AndNot, token.AndNotAssign)
			} else {
				tok = s.switch3(token.And, token.AndAssign, '&', token.LAnd)
			}
		case '|':
			tok = s.switch3(token.Or, token.OrAssign, '|', token.LOr)
		default:

			if ch != bom {
				s.error(s.file.Offset(pos),
					fmt.Sprintf("Geçersiz karakter %#U", ch))
			}
			insertSemi = s.insertSemi
			tok = token.Illegal
			literal = string(ch)
		}
	}
	if s.mode&DontInsertSemis == 0 {
		s.insertSemi = insertSemi
	}
	return
}

func (s *Scanner) next() {
	if s.readOffset < len(s.src) {
		s.offset = s.readOffset
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		r, w := rune(s.src[s.readOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "Geçersiz karakter NUL")
		case r >= utf8.RuneSelf:

			r, w = utf8.DecodeRune(s.src[s.readOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "Geçersiz UTF-8 encoding")
			} else if r == bom && s.offset > 0 {
				s.error(s.offset, "Geçersiz byte order mark")
			}
		}
		s.readOffset += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		if s.ch == '\n' {
			s.lineOffset = s.offset
			s.file.AddLine(s.offset)
		}
		s.ch = -1
	}
}

func (s *Scanner) peek() byte {
	if s.readOffset < len(s.src) {
		return s.src[s.readOffset]
	}
	return 0
}

func (s *Scanner) error(offset int, msg string) {
	if s.errorHandler != nil {
		s.errorHandler(s.file.Position(s.file.FileSetPos(offset)), msg)
	}
	s.errorCount++
}

func (s *Scanner) scanComment() string {

	offs := s.offset - 1
	var numCR int

	if s.ch == '/' {
		//-style comment

		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			if s.ch == '\r' {
				numCR++
			}
			s.next()
		}
		goto exit
	}

	/*-style comment */
	s.next()
	for s.ch >= 0 {
		ch := s.ch
		if ch == '\r' {
			numCR++
		}
		s.next()
		if ch == '*' && s.ch == '/' {
			s.next()
			goto exit
		}
	}

	s.error(offs, "Yorum kapatılmamış")

exit:
	lit := s.src[offs:s.offset]

	if numCR > 0 && len(lit) >= 2 && lit[1] == '/' && lit[len(lit)-1] == '\r' {
		lit = lit[:len(lit)-1]
		numCR--
	}
	if numCR > 0 {
		lit = StripCR(lit, lit[1] == '*')
	}
	return string(lit)
}

func (s *Scanner) findLineEnd() bool {

	defer func(offs int) {

		s.ch = '/'
		s.offset = offs
		s.readOffset = offs + 1
		s.next()
	}(s.offset - 1)

	for s.ch == '/' || s.ch == '*' {
		if s.ch == '/' {
			return true
		}
		s.next()
		for s.ch >= 0 {
			ch := s.ch
			if ch == '\n' {
				return true
			}
			s.next()
			if ch == '*' && s.ch == '/' {
				s.next()
				break
			}
		}
		s.skipWhitespace()
		if s.ch < 0 || s.ch == '\n' {
			return true
		}
		if s.ch != '/' {

			return false
		}
		s.next()
	}
	return false
}

func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanDigits(base int) {
	for s.ch == '_' || digitVal(s.ch) < base {
		s.next()
	}
}

func (s *Scanner) scanNumber() (token.Token, string) {
	offs := s.offset
	tok := token.Int
	base := 10

	switch {
	case s.ch == '0' && lower(s.peek()) == 'b':
		base = 2
		s.next()
		s.next()
	case s.ch == '0' && lower(s.peek()) == 'o':
		base = 8
		s.next()
		s.next()
	case s.ch == '0' && lower(s.peek()) == 'x':
		base = 16
		s.next()
		s.next()
	}

	s.scanDigits(base)

	if s.ch == '.' && (base == 10 || base == 16) {
		tok = token.Float
		s.next()
		s.scanDigits(base)
	}

	if s.ch == 'e' || s.ch == 'E' || s.ch == 'p' || s.ch == 'P' {
		tok = token.Float
		s.next()
		if s.ch == '-' || s.ch == '+' {
			s.next()
		}
		offs := s.offset
		s.scanDigits(10)
		if offs == s.offset {
			s.error(offs, "exponentnın ardından sayı yok")
		}
	}

	return tok, string(s.src[offs:s.offset])
}

func (s *Scanner) scanEscape(quote rune) bool {
	offs := s.offset

	var n int
	var base, max uint32
	switch s.ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		s.next()
		return true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.next()
		n, base, max = 2, 16, 255
	case 'u':
		s.next()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.next()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		msg := "bilinmeyen kaçış karakteri"
		if s.ch < 0 {
			msg = "kaçış karakteri yok"
		}
		s.error(offs, msg)
		return false
	}

	var x uint32
	for n > 0 {
		d := uint32(digitVal(s.ch))
		if d >= base {
			msg := fmt.Sprintf(
				"geçersiz karakter %#U", s.ch)
			if s.ch < 0 {
				msg = "kaçış karakteri yok"
			}
			s.error(s.offset, msg)
			return false
		}
		x = x*base + d
		s.next()
		n--
	}

	if x > max || 0xD800 <= x && x < 0xE000 {
		s.error(offs, "geçersiz Unicode kod noktası")
		return false
	}
	return true
}

func (s *Scanner) scanRune() string {
	offs := s.offset - 1

	valid := true
	n := 0
	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {

			if valid {
				s.error(offs, "karakter dizisi bitmedi")
				valid = false
			}
			break
		}
		s.next()
		if ch == '\'' {
			break
		}
		n++
		if ch == '\\' {
			if !s.scanEscape('\'') {
				valid = false
			}

		}
	}

	if valid && n != 1 {
		s.error(offs, "tek karakter bekleniyor")
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanString() string {
	offs := s.offset - 1

	for {
		ch := s.ch
		if ch == '\n' || ch < 0 {
			s.error(offs, "dize bitmedi")
			break
		}
		s.next()
		if ch == '"' {
			break
		}
		if ch == '\\' {
			s.scanEscape('"')
		}
	}
	return string(s.src[offs:s.offset])
}

func (s *Scanner) scanRawString() string {
	offs := s.offset - 1

	hasCR := false
	for {
		ch := s.ch
		if ch < 0 {
			s.error(offs, "dize bitmedi")
			break
		}

		s.next()

		if ch == '`' {
			break
		}

		if ch == '\r' {
			hasCR = true
		}
	}

	lit := s.src[offs:s.offset]
	if hasCR {
		lit = StripCR(lit, false)
	}
	return string(lit)
}

func StripCR(b []byte, comment bool) []byte {
	c := make([]byte, len(b))
	i := 0
	for j, ch := range b {

		if ch != '\r' || comment && i > len("/*") && c[i-1] == '*' &&
			j+1 < len(b) && b[j+1] == '/' {
			c[i] = ch
			i++
		}
	}
	return c[:i]
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' && !s.insertSemi ||
		s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) switch2(tok0, tok1 token.Token) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	return tok0
}

func (s *Scanner) switch3(
	tok0, tok1 token.Token,
	ch2 rune,
	tok2 token.Token,
) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	if s.ch == ch2 {
		s.next()
		return tok2
	}
	return tok0
}

func (s *Scanner) switch4(
	tok0, tok1 token.Token,
	ch2 rune,
	tok2, tok3 token.Token,
) token.Token {
	if s.ch == '=' {
		s.next()
		return tok1
	}
	if s.ch == ch2 {
		s.next()
		if s.ch == '=' {
			s.next()
			return tok3
		}
		return tok2
	}
	return tok0
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
		ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' ||
		ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16
}

func lower(c byte) byte {
	return c | ('x' - 'X')
}
