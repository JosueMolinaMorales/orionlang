package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/JosueMolinaMorales/monkeylang/internal/evaluator"
	"github.com/JosueMolinaMorales/monkeylang/internal/lexer"
	"github.com/JosueMolinaMorales/monkeylang/internal/object"
	"github.com/JosueMolinaMorales/monkeylang/internal/parser"
	"github.com/JosueMolinaMorales/monkeylang/internal/repl"
)

var (
	filePath = flag.String("path", ".", "The file path to the file that should be interpreted")
	runRepl  = flag.Bool("repl", false, "Run REPL for MonkeyLang")
)

func main() {
	flag.Parse()

	if *runRepl {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Hello %s! This is the Monkey programming language!\n", user.Username)

		fmt.Printf("Feel free to type in commands\n")

		repl.Start(os.Stdin, os.Stdout)
		return
	}

	if !strings.HasSuffix(*filePath, ".mkl") {
		log.Fatalf("File %s is not a MonkeyLang file", *filePath)
	}

	fileInfo, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	parser := parser.New(lexer.New(string(fileInfo)))
	program := parser.ParseProgram()

	if len(parser.Errors()) != 0 {
		repl.PrintParserErrors(os.Stdout, parser.Errors())
		return
	}

	evaluator.Eval(program, object.NewEnvironment())
}
