package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/onrirr/lokum"
	"github.com/onrirr/lokum/parser"
	"github.com/onrirr/lokum/stdlib"
)

const (
	sourceFileExt = ".lokum"
	replPrompt    = ">> "
)

var (
	compileOutput string
	showHelp      bool
	showVersion   bool
	resolvePath   bool
	version       = "beta"
)

func init() {
	flag.BoolVar(&showHelp, "yardım", false, "Yardım göster")
	flag.StringVar(&compileOutput, "o", "", "Compile çıkış dosyası")
	flag.BoolVar(&showVersion, "sürüm", false, "Sürümü göster")
	flag.BoolVar(&resolvePath, "resolve", false,
		"Importları çözümle.")
	flag.Parse()
}

func main() {
	if showHelp {
		doHelp()
		os.Exit(2)
	} else if showVersion {
		fmt.Println(version)
		return
	}

	modules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	inputFile := flag.Arg(0)
	if inputFile == "" {

		RunREPL(modules, os.Stdin, os.Stdout)
		return
	}

	inputData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"Dosyayı bulamadım: %s\n", err.Error())
		os.Exit(1)
	}

	inputFile, err = filepath.Abs(inputFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Dosyayı bulamadım: %s\n", err)
		os.Exit(1)
	}

	if len(inputData) > 1 && string(inputData[:2]) == "#!" {
		copy(inputData, "//")
	}

	if compileOutput != "" {
		err := CompileOnly(modules, inputData, inputFile,
			compileOutput)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else if filepath.Ext(inputFile) == sourceFileExt {
		err := CompileAndRun(modules, inputData, inputFile)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else {
		if err := RunCompiled(modules, inputData); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func CompileOnly(
	modules *lokum.ModuleMap,
	data []byte,
	inputFile, outputFile string,
) (err error) {
	bytecode, err := compileSrc(modules, data, inputFile)
	if err != nil {
		return
	}

	if outputFile == "" {
		outputFile = basename(inputFile) + ".out"
	}

	out, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = out.Close()
		} else {
			err = out.Close()
		}
	}()

	err = bytecode.Encode(out)
	if err != nil {
		return
	}
	fmt.Println(outputFile)
	return
}

func CompileAndRun(
	modules *lokum.ModuleMap,
	data []byte,
	inputFile string,
) (err error) {
	bytecode, err := compileSrc(modules, data, inputFile)
	if err != nil {
		return
	}

	machine := lokum.NewVM(bytecode, nil, -1)
	err = machine.Run()
	return
}

func RunCompiled(modules *lokum.ModuleMap, data []byte) (err error) {
	bytecode := &lokum.Bytecode{}
	err = bytecode.Decode(bytes.NewReader(data), modules)
	if err != nil {
		return
	}

	machine := lokum.NewVM(bytecode, nil, -1)
	err = machine.Run()
	return
}

func RunREPL(modules *lokum.ModuleMap, in io.Reader, out io.Writer) {
	stdin := bufio.NewScanner(in)
	fileSet := parser.NewFileSet()
	globals := make([]lokum.Object, lokum.GlobalsSize)
	symbolTable := lokum.NewSymbolTable()
	for idx, fn := range lokum.GetAllBuiltinFunctions() {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}

	symbol := symbolTable.Define("__repl_println__")
	globals[symbol.Index] = &lokum.UserFunction{
		Name: "println",
		Value: func(args ...lokum.Object) (ret lokum.Object, err error) {
			var printArgs []interface{}
			for _, arg := range args {
				if _, isUndefined := arg.(*lokum.Undefined); isUndefined {
					printArgs = append(printArgs, "<undefined>")
				} else {
					s, _ := lokum.ToString(arg)
					printArgs = append(printArgs, s)
				}
			}
			printArgs = append(printArgs, "\n")
			_, _ = fmt.Print(printArgs...)
			return
		},
	}

	var constants []lokum.Object
	for {
		_, _ = fmt.Fprint(out, replPrompt)
		scanned := stdin.Scan()
		if !scanned {
			return
		}

		line := stdin.Text()
		srcFile := fileSet.AddFile("repl", -1, len(line))
		p := parser.NewParser(srcFile, []byte(line), nil)
		file, err := p.ParseFile()
		if err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}

		file = addPrints(file)
		c := lokum.NewCompiler(srcFile, symbolTable, constants, modules, nil)
		if err := c.Compile(file); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}

		bytecode := c.Bytecode()
		machine := lokum.NewVM(bytecode, globals, -1)
		if err := machine.Run(); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}
		constants = bytecode.Constants
	}
}

func compileSrc(
	modules *lokum.ModuleMap,
	src []byte,
	inputFile string,
) (*lokum.Bytecode, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile(filepath.Base(inputFile), -1, len(src))

	p := parser.NewParser(srcFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := lokum.NewCompiler(srcFile, nil, nil, modules, nil)
	c.EnableFileImport(true)
	if resolvePath {
		c.SetImportDir(filepath.Dir(inputFile))
	}

	if err := c.Compile(file); err != nil {
		return nil, err
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()
	return bytecode, nil
}

func doHelp() {
	fmt.Println("...")
}

func addPrints(file *parser.File) *parser.File {
	var stmts []parser.Stmt
	for _, s := range file.Stmts {
		switch s := s.(type) {
		case *parser.ExprStmt:
			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{Name: "__repl_println__"},
					Args: []parser.Expr{s.Expr},
				},
			})
		case *parser.AssignStmt:
			stmts = append(stmts, s)

			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{
						Name: "__repl_println__",
					},
					Args: s.LHS,
				},
			})
		default:
			stmts = append(stmts, s)
		}
	}
	return &parser.File{
		InputFile: file.InputFile,
		Stmts:     stmts,
	}
}

func basename(s string) string {
	s = filepath.Base(s)
	n := strings.LastIndexByte(s, '.')
	if n > 0 {
		return s[:n]
	}
	return s
}
