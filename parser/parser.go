package parser

import (
	"Compiler/lexer"
	"fmt"
	"log"
	"os"
	"strings"
)

type Token = lexer.Token

var syntaxErrorFile *os.File
var currentToken Token

type ASTnode struct {
	Symbol   string
	Lexeme   string
	Type     string
	Children []ASTnode
}

func (n *ASTnode) addChild(child ASTnode) {
	n.Children = append(n.Children, child)
}

func parse(nonTerminal string) ASTnode {
	nonTerminalIndex, exists := nonTerminalMap[nonTerminal]
	if !exists {
		writeSyntaxError("Error: Invalid non-terminal " + nonTerminal) // Issue with LL1 table
		return ASTnode{}
	}
	currentTokenIndex, tokenExists := terminalMap[currentToken.Symbol]
	if !tokenExists {
		writeSyntaxError("Error: Invalid terminal " + currentToken.Lexeme) // Issue with LL1 or lexical
		return ASTnode{}
	}
	rule := ll1[nonTerminalIndex][currentTokenIndex]

	if rule == "" {
		fmt.Println("nonTerminal", nonTerminalIndex)
		fmt.Println("currentToken", currentTokenIndex)
		writeSyntaxError("Error: Unexpected " + currentToken.Lexeme) // Syntax error
		return ASTnode{}
	}

	node := ASTnode{Symbol: nonTerminal}

	rules := strings.Split(rule, " ")
	for _, symbol := range rules[1:] { // Discard first element (current non-terminal)
		_, exists := nonTerminalMap[symbol]
		if exists {
			child := parse(symbol) // Recursively parse non-terminals
			if child.Symbol == "" {
				return ASTnode{} // Return up call stack if error occured
			}
			node.addChild(child)
		} else if symbol == currentToken.Symbol {
			leafNode := ASTnode{Symbol: symbol, Lexeme: currentToken.Lexeme, Type: currentToken.Type}
			node.addChild(leafNode)
			currentToken = lexer.GetNextToken()
			if currentToken.Type == "ERROR" { // Invalid token, stop syntax parsing
				fmt.Println("Error: Invalid token", currentToken.Lexeme)
				return ASTnode{}
			}
			if currentToken.Type == "EOF" {
				if nonTerminal == "program" && symbol == "." {
					break // Valid end of program
				} else {
					writeSyntaxError("Error: Unexpected EOF")
					return ASTnode{}
				}
			}
		} else if symbol == "e" {
			continue
		} else {
			writeSyntaxError("Error: Invalid symbol " + symbol)
			return ASTnode{}
		}
	}
	return node
}

func writeSyntaxError(err string) {
	errorMessage := fmt.Sprintf("%s at line %d column %d", err, lexer.LineNumber, lexer.ColumnNumber)
	fmt.Println(errorMessage)
	fmt.Fprintln(syntaxErrorFile, errorMessage)
}

func StartParser() ASTnode {
	var err error
	syntaxErrorFile, err = os.Create("parser/syntax_errors.txt")
	if err != nil {
		log.Fatal("Error creating syntax_errors.txt")
	}
	currentToken = lexer.GetNextToken() // get first token
	root := parse("program")
	if root.Symbol == "" && !lexer.IsPanicMode {
		lexer.ErrorRecovery() // Init error recovery for lexer
		return ASTnode{}
	}
	return root
}

func CloseParserFiles() {
	defer syntaxErrorFile.Close()
}
