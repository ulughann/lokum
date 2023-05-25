package require

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/onrirr/lokum"
	"github.com/onrirr/lokum/parser"
	"github.com/onrirr/lokum/token"
)

func NoError(t *testing.T, err error, msg ...interface{}) {
	if err != nil {
		failExpectedActual(t, "Hata yok", err, msg...)
	}
}

func Error(t *testing.T, err error, msg ...interface{}) {
	if err == nil {
		failExpectedActual(t, "hata", err, msg...)
	}
}

func Nil(t *testing.T, v interface{}, msg ...interface{}) {
	if !isNil(v) {
		failExpectedActual(t, "nil", v, msg...)
	}
}

func True(t *testing.T, v bool, msg ...interface{}) {
	if !v {
		failExpectedActual(t, "doğru", v, msg...)
	}
}

func False(t *testing.T, v bool, msg ...interface{}) {
	if v {
		failExpectedActual(t, "yanlış", v, msg...)
	}
}

func NotNil(t *testing.T, v interface{}, msg ...interface{}) {
	if isNil(v) {
		failExpectedActual(t, "nil değil", v, msg...)
	}
}

func IsType(
	t *testing.T,
	expected, actual interface{},
	msg ...interface{},
) {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		failExpectedActual(t, reflect.TypeOf(expected),
			reflect.TypeOf(actual), msg...)
	}
}

func Equal(
	t *testing.T,
	expected, actual interface{},
	msg ...interface{},
) {
	if isNil(expected) {
		Nil(t, actual, "nil beklendi ama nil bulundu")
		return
	}
	NotNil(t, actual, "nil olmayan bir değer bekleniyor")
	IsType(t, expected, actual, msg...)

	switch expected := expected.(type) {
	case int:
		if expected != actual.(int) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case int64:
		if expected != actual.(int64) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case float64:
		if expected != actual.(float64) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case string:
		if expected != actual.(string) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case []byte:
		if !bytes.Equal(expected, actual.([]byte)) {
			failExpectedActual(t, string(expected),
				string(actual.([]byte)), msg...)
		}
	case []string:
		if !equalStringSlice(expected, actual.([]string)) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case []int:
		if !equalIntSlice(expected, actual.([]int)) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case bool:
		if expected != actual.(bool) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case rune:
		if expected != actual.(rune) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case *lokum.Symbol:
		if !equalSymbol(expected, actual.(*lokum.Symbol)) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case parser.Pos:
		if expected != actual.(parser.Pos) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case token.Token:
		if expected != actual.(token.Token) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case []lokum.Object:
		equalObjectSlice(t, expected, actual.([]lokum.Object), msg...)
	case *lokum.Int:
		Equal(t, expected.Value, actual.(*lokum.Int).Value, msg...)
	case *lokum.Float:
		Equal(t, expected.Value, actual.(*lokum.Float).Value, msg...)
	case *lokum.String:
		Equal(t, expected.Value, actual.(*lokum.String).Value, msg...)
	case *lokum.Char:
		Equal(t, expected.Value, actual.(*lokum.Char).Value, msg...)
	case *lokum.Bool:
		if expected != actual {
			failExpectedActual(t, expected, actual, msg...)
		}
	case *lokum.Array:
		equalObjectSlice(t, expected.Value,
			actual.(*lokum.Array).Value, msg...)
	case *lokum.ImmutableArray:
		equalObjectSlice(t, expected.Value,
			actual.(*lokum.ImmutableArray).Value, msg...)
	case *lokum.Bytes:
		if !bytes.Equal(expected.Value, actual.(*lokum.Bytes).Value) {
			failExpectedActual(t, string(expected.Value),
				string(actual.(*lokum.Bytes).Value), msg...)
		}
	case *lokum.Map:
		equalObjectMap(t, expected.Value,
			actual.(*lokum.Map).Value, msg...)
	case *lokum.ImmutableMap:
		equalObjectMap(t, expected.Value,
			actual.(*lokum.ImmutableMap).Value, msg...)
	case *lokum.CompiledFunction:
		equalCompiledFunction(t, expected,
			actual.(*lokum.CompiledFunction), msg...)
	case *lokum.Undefined:
		if expected != actual {
			failExpectedActual(t, expected, actual, msg...)
		}
	case *lokum.Error:
		Equal(t, expected.Value, actual.(*lokum.Error).Value, msg...)
	case lokum.Object:
		if !expected.Equals(actual.(lokum.Object)) {
			failExpectedActual(t, expected, actual, msg...)
		}
	case *parser.SourceFileSet:
		equalFileSet(t, expected, actual.(*parser.SourceFileSet), msg...)
	case *parser.SourceFile:
		Equal(t, expected.Name, actual.(*parser.SourceFile).Name, msg...)
		Equal(t, expected.Base, actual.(*parser.SourceFile).Base, msg...)
		Equal(t, expected.Size, actual.(*parser.SourceFile).Size, msg...)
		True(t, equalIntSlice(expected.Lines,
			actual.(*parser.SourceFile).Lines), msg...)
	case error:
		if expected != actual.(error) {
			failExpectedActual(t, expected, actual, msg...)
		}
	default:
		panic(fmt.Errorf("Tip iplemente edilmemiş: %T", expected))
	}
}

func Fail(t *testing.T, msg ...interface{}) {
	t.Logf("\nHata:\n\t%s\n%s", strings.Join(errorTrace(), "\n\t"),
		message(msg...))
	t.Fail()
}

func failExpectedActual(
	t *testing.T,
	expected, actual interface{},
	msg ...interface{},
) {
	var addMsg string
	if len(msg) > 0 {
		addMsg = "\nMesaj:  " + message(msg...)
	}

	t.Logf("\nHata:\n\t%s\nBeklenen: %v\nBulunan:   %v%s",
		strings.Join(errorTrace(), "\n\t"),
		expected, actual,
		addMsg)
	t.FailNow()
}

func message(formatArgs ...interface{}) string {
	var format string
	var args []interface{}
	if len(formatArgs) > 0 {
		format = formatArgs[0].(string)
	}
	if len(formatArgs) > 1 {
		args = formatArgs[1:]
	}
	return fmt.Sprintf(format, args...)
}

func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalSymbol(a, b *lokum.Symbol) bool {
	return a.Name == b.Name &&
		a.Index == b.Index &&
		a.Scope == b.Scope
}

func equalObjectSlice(
	t *testing.T,
	expected, actual []lokum.Object,
	msg ...interface{},
) {
	Equal(t, len(expected), len(actual), msg...)
	for i := 0; i < len(expected); i++ {
		Equal(t, expected[i], actual[i], msg...)
	}
}

func equalFileSet(
	t *testing.T,
	expected, actual *parser.SourceFileSet,
	msg ...interface{},
) {
	Equal(t, len(expected.Files), len(actual.Files), msg...)
	for i, f := range expected.Files {
		Equal(t, f, actual.Files[i], msg...)
	}
	Equal(t, expected.Base, actual.Base)
	Equal(t, expected.LastFile, actual.LastFile)
}

func equalObjectMap(
	t *testing.T,
	expected, actual map[string]lokum.Object,
	msg ...interface{},
) {
	Equal(t, len(expected), len(actual), msg...)
	for key, expectedVal := range expected {
		actualVal := actual[key]
		Equal(t, expectedVal, actualVal, msg...)
	}
}

func equalCompiledFunction(
	t *testing.T,
	expected, actual lokum.Object,
	msg ...interface{},
) {
	expectedT := expected.(*lokum.CompiledFunction)
	actualT := actual.(*lokum.CompiledFunction)
	Equal(t,
		lokum.FormatInstructions(expectedT.Instructions, 0),
		lokum.FormatInstructions(actualT.Instructions, 0), msg...)
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	kind := value.Kind()
	return kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil()
}

func errorTrace() []string {
	var pc uintptr
	file := ""
	line := 0
	var ok bool
	name := ""

	var callers []string
	for i := 0; ; i++ {
		pc, file, line, ok = runtime.Caller(i)
		if !ok {
			break
		}

		if file == "<autogenerated>" {
			break
		}

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		name = f.Name()

		if name == "testing.tRunner" {
			break
		}

		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
		if len(parts) > 1 {
			dir := parts[len(parts)-2]
			if dir != "require" ||
				file == "mock_test.go" {
				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
			}
		}

		segments := strings.Split(name, ".")
		name = segments[len(segments)-1]
		if isTest(name, "Test") ||
			isTest(name, "Benchmark") ||
			isTest(name, "Example") {
			break
		}
	}
	return callers
}

func isTest(name, prefix string) bool {
	if !strings.HasPrefix(name, prefix) {
		return false
	}
	if len(name) == len(prefix) {
		return true
	}
	r, _ := utf8.DecodeRuneInString(name[len(prefix):])
	return !unicode.IsLower(r)
}
