package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/JosueMolinaMorales/orionlang/internal/compiler"
	"github.com/JosueMolinaMorales/orionlang/internal/evaluator"
	"github.com/JosueMolinaMorales/orionlang/internal/lexer"
	"github.com/JosueMolinaMorales/orionlang/internal/object"
	"github.com/JosueMolinaMorales/orionlang/internal/parser"
	"github.com/JosueMolinaMorales/orionlang/internal/vm"
)

// PROMPT is the prompt of the REPL
const PROMPT = ">> "

// Start starts the REPL
func Start(in io.Reader, out io.Writer, useInterpreter bool) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}

		if useInterpreter {
			evaluated := evaluator.Eval(program, env)
			if evaluated != nil {
				io.WriteString(out, evaluated.Inspect())
				io.WriteString(out, "\n")
			}
			return
		}
		// Compiler
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Executing bytecode failed:\n %s\n", err)
			continue
		}

		stackTop := machine.StackTop()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

	}
}

func PrintParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Woops! We ran into some errors!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
