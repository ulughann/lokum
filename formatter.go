package lokum

import (
	"strconv"
	"sync"
	"unicode/utf8"
)

const (
	commaSpaceString  = ", "
	nilParenString    = "(nil)"
	percentBangString = "%!"
	missingString     = "(MISSING)"
	badIndexString    = "(BADINDEX)"
	extraString       = "%!(EXTRA "
	badWidthString    = "%!(BADWIDTH)"
	badPrecString     = "%!(BADPREC)"
	noVerbString      = "%!(NOVERB)"
)

const (
	ldigits = "0123456789abcdefx"
	udigits = "0123456789ABCDEFX"
)

const (
	signed   = true
	unsigned = false
)

type fmtFlags struct {
	widPresent  bool
	precPresent bool
	minus       bool
	plus        bool
	sharp       bool
	space       bool
	zero        bool

	plusV  bool
	sharpV bool

	inDetail    bool
	needNewline bool
	needColon   bool
}

type formatter struct {
	buf *fmtbuf

	fmtFlags

	wid  int
	prec int

	intbuf [68]byte
}

func (f *formatter) clearFlags() {
	f.fmtFlags = fmtFlags{}
}

func (f *formatter) init(buf *fmtbuf) {
	f.buf = buf
	f.clearFlags()
}

func (f *formatter) writePadding(n int) {
	if n <= 0 {
		return
	}
	buf := *f.buf
	oldLen := len(buf)
	newLen := oldLen + n

	if newLen > MaxStringLen {
		panic(ErrStringLimit)
	}

	if newLen > cap(buf) {
		buf = make(fmtbuf, cap(buf)*2+n)
		copy(buf, *f.buf)
	}

	padByte := byte(' ')
	if f.zero {
		padByte = byte('0')
	}

	padding := buf[oldLen:newLen]
	for i := range padding {
		padding[i] = padByte
	}
	*f.buf = buf[:newLen]
}

func (f *formatter) pad(b []byte) {
	if !f.widPresent || f.wid == 0 {
		f.buf.Write(b)
		return
	}
	width := f.wid - utf8.RuneCount(b)
	if !f.minus {

		f.writePadding(width)
		f.buf.Write(b)
	} else {

		f.buf.Write(b)
		f.writePadding(width)
	}
}

func (f *formatter) padString(s string) {
	if !f.widPresent || f.wid == 0 {
		f.buf.WriteString(s)
		return
	}
	width := f.wid - utf8.RuneCountInString(s)
	if !f.minus {

		f.writePadding(width)
		f.buf.WriteString(s)
	} else {

		f.buf.WriteString(s)
		f.writePadding(width)
	}
}

func (f *formatter) fmtBoolean(v bool) {
	if v {
		f.padString("true")
	} else {
		f.padString("false")
	}
}

func (f *formatter) fmtUnicode(u uint64) {
	buf := f.intbuf[0:]

	prec := 4
	if f.precPresent && f.prec > 4 {
		prec = f.prec

		width := 2 + prec + 2 + utf8.UTFMax + 1
		if width > len(buf) {
			buf = make([]byte, width)
		}
	}

	i := len(buf)

	if f.sharp && u <= utf8.MaxRune && strconv.IsPrint(rune(u)) {
		i--
		buf[i] = '\''
		i -= utf8.RuneLen(rune(u))
		utf8.EncodeRune(buf[i:], rune(u))
		i--
		buf[i] = '\''
		i--
		buf[i] = ' '
	}

	for u >= 16 {
		i--
		buf[i] = udigits[u&0xF]
		prec--
		u >>= 4
	}
	i--
	buf[i] = udigits[u]
	prec--

	for prec > 0 {
		i--
		buf[i] = '0'
		prec--
	}

	i--
	buf[i] = '+'
	i--
	buf[i] = 'U'

	oldZero := f.zero
	f.zero = false
	f.pad(buf[i:])
	f.zero = oldZero
}

func (f *formatter) fmtInteger(
	u uint64,
	base int,
	isSigned bool,
	verb rune,
	digits string,
) {
	negative := isSigned && int64(u) < 0
	if negative {
		u = -u
	}

	buf := f.intbuf[0:]

	if f.widPresent || f.precPresent {

		width := 3 + f.wid + f.prec
		if width > len(buf) {

			buf = make([]byte, width)
		}
	}

	prec := 0
	if f.precPresent {
		prec = f.prec

		if prec == 0 && u == 0 {
			oldZero := f.zero
			f.zero = false
			f.writePadding(f.wid)
			f.zero = oldZero
			return
		}
	} else if f.zero && f.widPresent {
		prec = f.wid
		if negative || f.plus || f.space {
			prec--
		}
	}

	i := len(buf)

	switch base {
	case 10:
		for u >= 10 {
			i--
			next := u / 10
			buf[i] = byte('0' + u - next*10)
			u = next
		}
	case 16:
		for u >= 16 {
			i--
			buf[i] = digits[u&0xF]
			u >>= 4
		}
	case 8:
		for u >= 8 {
			i--
			buf[i] = byte('0' + u&7)
			u >>= 3
		}
	case 2:
		for u >= 2 {
			i--
			buf[i] = byte('0' + u&1)
			u >>= 1
		}
	default:
		panic("fmt: unknown base; can't happen")
	}
	i--
	buf[i] = digits[u]
	for i > 0 && prec > len(buf)-i {
		i--
		buf[i] = '0'
	}

	if f.sharp {
		switch base {
		case 2:

			i--
			buf[i] = 'b'
			i--
			buf[i] = '0'
		case 8:
			if buf[i] != '0' {
				i--
				buf[i] = '0'
			}
		case 16:

			i--
			buf[i] = digits[16]
			i--
			buf[i] = '0'
		}
	}
	if verb == 'O' {
		i--
		buf[i] = 'o'
		i--
		buf[i] = '0'
	}

	if negative {
		i--
		buf[i] = '-'
	} else if f.plus {
		i--
		buf[i] = '+'
	} else if f.space {
		i--
		buf[i] = ' '
	}

	oldZero := f.zero
	f.zero = false
	f.pad(buf[i:])
	f.zero = oldZero
}

func (f *formatter) truncateString(s string) string {
	if f.precPresent {
		n := f.prec
		for i := range s {
			n--
			if n < 0 {
				return s[:i]
			}
		}
	}
	return s
}

func (f *formatter) truncate(b []byte) []byte {
	if f.precPresent {
		n := f.prec
		for i := 0; i < len(b); {
			n--
			if n < 0 {
				return b[:i]
			}
			wid := 1
			if b[i] >= utf8.RuneSelf {
				_, wid = utf8.DecodeRune(b[i:])
			}
			i += wid
		}
	}
	return b
}

func (f *formatter) fmtS(s string) {
	s = f.truncateString(s)
	f.padString(s)
}

func (f *formatter) fmtBs(b []byte) {
	b = f.truncate(b)
	f.pad(b)
}

func (f *formatter) fmtSbx(s string, b []byte, digits string) {
	length := len(b)
	if b == nil {

		length = len(s)
	}

	if f.precPresent && f.prec < length {
		length = f.prec
	}

	width := 2 * length
	if width > 0 {
		if f.space {

			if f.sharp {
				width *= 2
			}

			width += length - 1
		} else if f.sharp {

			width += 2
		}
	} else {
		if f.widPresent {
			f.writePadding(f.wid)
		}
		return
	}

	if f.widPresent && f.wid > width && !f.minus {
		f.writePadding(f.wid - width)
	}

	buf := *f.buf
	if f.sharp {

		buf = append(buf, '0', digits[16])
	}
	var c byte
	for i := 0; i < length; i++ {
		if f.space && i > 0 {

			buf = append(buf, ' ')
			if f.sharp {

				buf = append(buf, '0', digits[16])
			}
		}
		if b != nil {
			c = b[i]
		} else {
			c = s[i]
		}

		buf = append(buf, digits[c>>4], digits[c&0xF])
	}
	*f.buf = buf

	if f.widPresent && f.wid > width && f.minus {
		f.writePadding(f.wid - width)
	}
}

func (f *formatter) fmtSx(s, digits string) {
	f.fmtSbx(s, nil, digits)
}

func (f *formatter) fmtBx(b []byte, digits string) {
	f.fmtSbx("", b, digits)
}

func (f *formatter) fmtQ(s string) {
	s = f.truncateString(s)
	if f.sharp && strconv.CanBackquote(s) {
		f.padString("`" + s + "`")
		return
	}
	buf := f.intbuf[:0]
	if f.plus {
		f.pad(strconv.AppendQuoteToASCII(buf, s))
	} else {
		f.pad(strconv.AppendQuote(buf, s))
	}
}

func (f *formatter) fmtC(c uint64) {
	r := rune(c)
	if c > utf8.MaxRune {
		r = utf8.RuneError
	}
	buf := f.intbuf[:0]
	w := utf8.EncodeRune(buf[:utf8.UTFMax], r)
	f.pad(buf[:w])
}

func (f *formatter) fmtQc(c uint64) {
	r := rune(c)
	if c > utf8.MaxRune {
		r = utf8.RuneError
	}
	buf := f.intbuf[:0]
	if f.plus {
		f.pad(strconv.AppendQuoteRuneToASCII(buf, r))
	} else {
		f.pad(strconv.AppendQuoteRune(buf, r))
	}
}

func (f *formatter) fmtFloat(v float64, size int, verb rune, prec int) {

	if f.precPresent {
		prec = f.prec
	}

	num := strconv.AppendFloat(f.intbuf[:1], v, byte(verb), prec, size)
	if num[1] == '-' || num[1] == '+' {
		num = num[1:]
	} else {
		num[0] = '+'
	}

	if f.space && num[0] == '+' && !f.plus {
		num[0] = ' '
	}

	if num[1] == 'I' || num[1] == 'N' {
		oldZero := f.zero
		f.zero = false

		if num[1] == 'N' && !f.space && !f.plus {
			num = num[1:]
		}
		f.pad(num)
		f.zero = oldZero
		return
	}

	if f.sharp && verb != 'b' {
		digits := 0
		switch verb {
		case 'v', 'g', 'G', 'x':
			digits = prec

			if digits == -1 {
				digits = 6
			}
		}

		var tailBuf [6]byte
		tail := tailBuf[:0]

		hasDecimalPoint := false

		for i := 1; i < len(num); i++ {
			switch num[i] {
			case '.':
				hasDecimalPoint = true
			case 'p', 'P':
				tail = append(tail, num[i:]...)
				num = num[:i]
			case 'e', 'E':
				if verb != 'x' && verb != 'X' {
					tail = append(tail, num[i:]...)
					num = num[:i]
					break
				}
				fallthrough
			default:
				digits--
			}
		}
		if !hasDecimalPoint {
			num = append(num, '.')
		}
		for digits > 0 {
			num = append(num, '0')
			digits--
		}
		num = append(num, tail...)
	}

	if f.plus || num[0] != '+' {

		if f.zero && f.widPresent && f.wid > len(num) {
			f.buf.WriteSingleByte(num[0])
			f.writePadding(f.wid - len(num))
			f.buf.Write(num[1:])
			return
		}
		f.pad(num)
		return
	}

	f.pad(num[1:])
}

type fmtbuf []byte

func (b *fmtbuf) Write(p []byte) {
	if len(*b)+len(p) > MaxStringLen {
		panic(ErrStringLimit)
	}

	*b = append(*b, p...)
}

func (b *fmtbuf) WriteString(s string) {
	if len(*b)+len(s) > MaxStringLen {
		panic(ErrStringLimit)
	}

	*b = append(*b, s...)
}

func (b *fmtbuf) WriteSingleByte(c byte) {
	if len(*b) >= MaxStringLen {
		panic(ErrStringLimit)
	}

	*b = append(*b, c)
}

func (b *fmtbuf) WriteRune(r rune) {
	if len(*b)+utf8.RuneLen(r) > MaxStringLen {
		panic(ErrStringLimit)
	}

	if r < utf8.RuneSelf {
		*b = append(*b, byte(r))
		return
	}

	b2 := *b
	n := len(b2)
	for n+utf8.UTFMax > cap(b2) {
		b2 = append(b2, 0)
	}
	w := utf8.EncodeRune(b2[n:n+utf8.UTFMax], r)
	*b = b2[:n+w]
}

type pp struct {
	buf fmtbuf

	arg Object

	fmt formatter

	reordered bool

	goodArgNum bool

	erroring bool
}

var ppFree = sync.Pool{
	New: func() interface{} { return new(pp) },
}

func newPrinter() *pp {
	p := ppFree.Get().(*pp)
	p.erroring = false
	p.fmt.init(&p.buf)
	return p
}

func (p *pp) free() {

	//

	if cap(p.buf) > 64<<10 {
		return
	}

	p.buf = p.buf[:0]
	p.arg = nil
	ppFree.Put(p)
}

func (p *pp) Width() (wid int, ok bool) {
	return p.fmt.wid, p.fmt.widPresent
}

func (p *pp) Precision() (prec int, ok bool) {
	return p.fmt.prec, p.fmt.precPresent
}

func (p *pp) Flag(b int) bool {
	switch b {
	case '-':
		return p.fmt.minus
	case '+':
		return p.fmt.plus || p.fmt.plusV
	case '#':
		return p.fmt.sharp || p.fmt.sharpV
	case ' ':
		return p.fmt.space
	case '0':
		return p.fmt.zero
	}
	return false
}

func (p *pp) Write(b []byte) (ret int, err error) {
	p.buf.Write(b)
	return len(b), nil
}

func (p *pp) WriteString(s string) (ret int, err error) {
	p.buf.WriteString(s)
	return len(s), nil
}

func (p *pp) WriteRune(r rune) (ret int, err error) {
	p.buf.WriteRune(r)
	return utf8.RuneLen(r), nil
}

func (p *pp) WriteSingleByte(c byte) (ret int, err error) {
	p.buf.WriteSingleByte(c)
	return 1, nil
}

func tooLarge(x int) bool {
	const max int = 1e6
	return x > max || x < -max
}

func parsenum(s string, start, end int) (num int, isnum bool, newi int) {
	if start >= end {
		return 0, false, end
	}
	for newi = start; newi < end && '0' <= s[newi] && s[newi] <= '9'; newi++ {
		if tooLarge(num) {
			return 0, false, end
		}
		num = num*10 + int(s[newi]-'0')
		isnum = true
	}
	return
}

func (p *pp) badVerb(verb rune) {
	p.erroring = true
	_, _ = p.WriteString(percentBangString)
	_, _ = p.WriteRune(verb)
	_, _ = p.WriteSingleByte('(')
	switch {
	case p.arg != nil:
		_, _ = p.WriteString(p.arg.String())
		_, _ = p.WriteSingleByte('=')
		p.printArg(p.arg, 'v')
	default:
		_, _ = p.WriteString(UndefinedValue.String())
	}
	_, _ = p.WriteSingleByte(')')
	p.erroring = false
}

func (p *pp) fmtBool(v bool, verb rune) {
	switch verb {
	case 't', 'v':
		p.fmt.fmtBoolean(v)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmt0x64(v uint64, leading0x bool) {
	sharp := p.fmt.sharp
	p.fmt.sharp = leading0x
	p.fmt.fmtInteger(v, 16, unsigned, 'v', ldigits)
	p.fmt.sharp = sharp
}

func (p *pp) fmtInteger(v uint64, isSigned bool, verb rune) {
	switch verb {
	case 'v':
		if p.fmt.sharpV && !isSigned {
			p.fmt0x64(v, true)
		} else {
			p.fmt.fmtInteger(v, 10, isSigned, verb, ldigits)
		}
	case 'd':
		p.fmt.fmtInteger(v, 10, isSigned, verb, ldigits)
	case 'b':
		p.fmt.fmtInteger(v, 2, isSigned, verb, ldigits)
	case 'o', 'O':
		p.fmt.fmtInteger(v, 8, isSigned, verb, ldigits)
	case 'x':
		p.fmt.fmtInteger(v, 16, isSigned, verb, ldigits)
	case 'X':
		p.fmt.fmtInteger(v, 16, isSigned, verb, udigits)
	case 'c':
		p.fmt.fmtC(v)
	case 'q':
		if v <= utf8.MaxRune {
			p.fmt.fmtQc(v)
		} else {
			p.badVerb(verb)
		}
	case 'U':
		p.fmt.fmtUnicode(v)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmtFloat(v float64, size int, verb rune) {
	switch verb {
	case 'v':
		p.fmt.fmtFloat(v, size, 'g', -1)
	case 'b', 'g', 'G', 'x', 'X':
		p.fmt.fmtFloat(v, size, verb, -1)
	case 'f', 'e', 'E':
		p.fmt.fmtFloat(v, size, verb, 6)
	case 'F':
		p.fmt.fmtFloat(v, size, 'f', 6)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmtString(v string, verb rune) {
	switch verb {
	case 'v':
		if p.fmt.sharpV {
			p.fmt.fmtQ(v)
		} else {
			p.fmt.fmtS(v)
		}
	case 's':
		p.fmt.fmtS(v)
	case 'x':
		p.fmt.fmtSx(v, ldigits)
	case 'X':
		p.fmt.fmtSx(v, udigits)
	case 'q':
		p.fmt.fmtQ(v)
	default:
		p.badVerb(verb)
	}
}

func (p *pp) fmtBytes(v []byte, verb rune, typeString string) {
	switch verb {
	case 'v', 'd':
		if p.fmt.sharpV {
			_, _ = p.WriteString(typeString)
			if v == nil {
				_, _ = p.WriteString(nilParenString)
				return
			}
			_, _ = p.WriteSingleByte('{')
			for i, c := range v {
				if i > 0 {
					_, _ = p.WriteString(commaSpaceString)
				}
				p.fmt0x64(uint64(c), true)
			}
			_, _ = p.WriteSingleByte('}')
		} else {
			_, _ = p.WriteSingleByte('[')
			for i, c := range v {
				if i > 0 {
					_, _ = p.WriteSingleByte(' ')
				}
				p.fmt.fmtInteger(uint64(c), 10, unsigned, verb, ldigits)
			}
			_, _ = p.WriteSingleByte(']')
		}
	case 's':
		p.fmt.fmtBs(v)
	case 'x':
		p.fmt.fmtBx(v, ldigits)
	case 'X':
		p.fmt.fmtBx(v, udigits)
	case 'q':
		p.fmt.fmtQ(string(v))
	}
}

func (p *pp) printArg(arg Object, verb rune) {
	p.arg = arg

	if arg == nil {
		arg = UndefinedValue
	}

	switch verb {
	case 'T':
		p.fmt.fmtS(arg.TypeName())
		return
	case 'v':
		p.fmt.fmtS(arg.String())
		return
	}

	switch f := arg.(type) {
	case *Bool:
		p.fmtBool(!f.IsFalsy(), verb)
	case *Float:
		p.fmtFloat(f.Value, 64, verb)
	case *Int:
		p.fmtInteger(uint64(f.Value), signed, verb)
	case *String:
		p.fmtString(f.Value, verb)
	case *Bytes:
		p.fmtBytes(f.Value, verb, "[]byte")
	default:
		p.fmtString(f.String(), verb)
	}
}

func intFromArg(a []Object, argNum int) (num int, isInt bool, newArgNum int) {
	newArgNum = argNum
	if argNum < len(a) {
		var num64 int64
		num64, isInt = ToInt64(a[argNum])
		num = int(num64)
		newArgNum = argNum + 1
		if tooLarge(num) {
			num = 0
			isInt = false
		}
	}
	return
}

func parseArgNumber(format string) (index int, wid int, ok bool) {

	if len(format) < 3 {
		return 0, 1, false
	}

	for i := 1; i < len(format); i++ {
		if format[i] == ']' {
			width, ok, newi := parsenum(format, 1, i)
			if !ok || newi != i {
				return 0, i + 1, false
			}

			return width - 1, i + 1, true
		}
	}
	return 0, 1, false
}

func (p *pp) argNumber(
	argNum int,
	format string,
	i int,
	numArgs int,
) (newArgNum, newi int, found bool) {
	if len(format) <= i || format[i] != '[' {
		return argNum, i, false
	}
	p.reordered = true
	index, wid, ok := parseArgNumber(format[i:])
	if ok && 0 <= index && index < numArgs {
		return index, i + wid, true
	}
	p.goodArgNum = false
	return argNum, i + wid, ok
}

func (p *pp) badArgNum(verb rune) {
	_, _ = p.WriteString(percentBangString)
	_, _ = p.WriteRune(verb)
	_, _ = p.WriteString(badIndexString)
}

func (p *pp) missingArg(verb rune) {
	_, _ = p.WriteString(percentBangString)
	_, _ = p.WriteRune(verb)
	_, _ = p.WriteString(missingString)
}

func (p *pp) doFormat(format string, a []Object) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && e == ErrStringLimit {
				err = e
				return
			}
			panic(r)
		}
	}()

	end := len(format)
	argNum := 0
	afterIndex := false
	p.reordered = false
formatLoop:
	for i := 0; i < end; {
		p.goodArgNum = true
		lasti := i
		for i < end && format[i] != '%' {
			i++
		}
		if i > lasti {
			_, _ = p.WriteString(format[lasti:i])
		}
		if i >= end {

			break
		}

		i++

		p.fmt.clearFlags()
	simpleFormat:
		for ; i < end; i++ {
			c := format[i]
			switch c {
			case '#':
				p.fmt.sharp = true
			case '0':

				p.fmt.zero = !p.fmt.minus
			case '+':
				p.fmt.plus = true
			case '-':
				p.fmt.minus = true
				p.fmt.zero = false
			case ' ':
				p.fmt.space = true
			default:

				if 'a' <= c && c <= 'z' && argNum < len(a) {
					if c == 'v' {

						p.fmt.sharpV = p.fmt.sharp
						p.fmt.sharp = false

						p.fmt.plusV = p.fmt.plus
						p.fmt.plus = false
					}
					p.printArg(a[argNum], rune(c))
					argNum++
					i++
					continue formatLoop
				}

				break simpleFormat
			}
		}

		argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))

		if i < end && format[i] == '*' {
			i++
			p.fmt.wid, p.fmt.widPresent, argNum = intFromArg(a, argNum)

			if !p.fmt.widPresent {
				_, _ = p.WriteString(badWidthString)
			}

			if p.fmt.wid < 0 {
				p.fmt.wid = -p.fmt.wid
				p.fmt.minus = true
				p.fmt.zero = false
			}
			afterIndex = false
		} else {
			p.fmt.wid, p.fmt.widPresent, i = parsenum(format, i, end)
			if afterIndex && p.fmt.widPresent {
				p.goodArgNum = false
			}
		}

		if i+1 < end && format[i] == '.' {
			i++
			if afterIndex {
				p.goodArgNum = false
			}
			argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))
			if i < end && format[i] == '*' {
				i++
				p.fmt.prec, p.fmt.precPresent, argNum = intFromArg(a, argNum)

				if p.fmt.prec < 0 {
					p.fmt.prec = 0
					p.fmt.precPresent = false
				}
				if !p.fmt.precPresent {
					_, _ = p.WriteString(badPrecString)
				}
				afterIndex = false
			} else {
				p.fmt.prec, p.fmt.precPresent, i = parsenum(format, i, end)
				if !p.fmt.precPresent {
					p.fmt.prec = 0
					p.fmt.precPresent = true
				}
			}
		}

		if !afterIndex {
			argNum, i, afterIndex = p.argNumber(argNum, format, i, len(a))
		}

		if i >= end {
			_, _ = p.WriteString(noVerbString)
			break
		}

		verb, size := rune(format[i]), 1
		if verb >= utf8.RuneSelf {
			verb, size = utf8.DecodeRuneInString(format[i:])
		}
		i += size

		switch {
		case verb == '%':

			_, _ = p.WriteSingleByte('%')
		case !p.goodArgNum:
			p.badArgNum(verb)
		case argNum >= len(a):

			p.missingArg(verb)
		case verb == 'v':

			p.fmt.sharpV = p.fmt.sharp
			p.fmt.sharp = false

			p.fmt.plusV = p.fmt.plus
			p.fmt.plus = false
			fallthrough
		default:
			p.printArg(a[argNum], verb)
			argNum++
		}
	}

	if !p.reordered && argNum < len(a) {
		p.fmt.clearFlags()
		_, _ = p.WriteString(extraString)
		for i, arg := range a[argNum:] {
			if i > 0 {
				_, _ = p.WriteString(commaSpaceString)
			}
			if arg == nil {
				_, _ = p.WriteString(UndefinedValue.String())
			} else {
				_, _ = p.WriteString(arg.TypeName())
				_, _ = p.WriteSingleByte('=')
				p.printArg(arg, 'v')
			}
		}
		_, _ = p.WriteSingleByte(')')
	}

	return nil
}

func Format(format string, a ...Object) (string, error) {
	p := newPrinter()
	err := p.doFormat(format, a)
	s := string(p.buf)
	p.free()

	return s, err
}
