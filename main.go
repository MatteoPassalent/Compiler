package main

import (
	"Compiler/codegen"
	"Compiler/lexer"
	"Compiler/parser"
	"Compiler/semantic"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <source_file>")
		os.Exit(1)
	}

	sourceFile := os.Args[1]
	file, err := os.Open(sourceFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", sourceFile, err)
		os.Exit(1)
	}
	defer file.Close()

	lexer.InitLexerFiles(file)
	root := parser.StartParser()
	semantic.SemanticAnalysis(&root)
	codegen.GenerateCode(&root)
	lexer.CloseLexerFiles()
	parser.CloseParserFiles()
}
