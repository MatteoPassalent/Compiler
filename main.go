package main

import (
	"Compiler/lexer"
	"Compiler/parser"
)

func main() {
	lexer.InitLexerFiles()
	parser.StartParser()

	lexer.CloseLexerFiles()
	parser.CloseParserFiles()
}
