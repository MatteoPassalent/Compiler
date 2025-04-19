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
	codegen.GenerateTAC(&root)
	lexer.CloseLexerFiles()
	parser.CloseParserFiles()

	// Print Error Messages
	printErrorFile("lexer/lexical_errors.txt", "Lexer Errors:")
	printErrorFile("parser/syntax_errors.txt", "Syntax Errors:")
	printErrorFile("semantic/semantic_errors.txt", "Semantic Errors:")
}

func printErrorFile(filePath, errorType string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening %s file: %v\n", errorType, err)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file info for %s: %v\n", errorType, err)
		return
	}

	if stat.Size() != 0 {
		fmt.Println(errorType)
		file.Seek(0, 0)
		file.WriteTo(os.Stdout)
	}
}
