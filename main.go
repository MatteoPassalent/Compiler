package main

import (
	"Compiler/lexer"
	"Compiler/parser"
	"Compiler/semantic"
)

func main() {
	lexer.InitLexerFiles()
	root := parser.StartParser()
	semantic.SemanticAnalysis(&root)
	lexer.CloseLexerFiles()
	parser.CloseParserFiles()
}
