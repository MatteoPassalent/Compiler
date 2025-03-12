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
	symbol     string
	children   []ASTnode
	childCount int
}

func (n *ASTnode) init(symbol string) {
	n.symbol = symbol
	n.children = make([]ASTnode, 0)
	n.childCount = 0
}

func (n *ASTnode) addChild(child ASTnode) {
	n.children = append(n.children, child)
	n.childCount++
}

func parse(nonTerminal string) ASTnode {
	nonTerminalIndex, exists := nonTerminalMap[nonTerminal]
	if !exists {
		writeSyntaxError("Error: Invalid non-terminal " + nonTerminal) // Issue with LL1 table
		return ASTnode{}
	}
	currentTokenIndex, tokenExists := terminalMap[currentToken.Value]
	if !tokenExists {
		writeSyntaxError("Error: Invalid terminal " + currentToken.Value) // Issue with LL1 or lexical
		return ASTnode{}
	}
	rule := ll1[nonTerminalIndex][currentTokenIndex]
	if rule == "" {
		writeSyntaxError("Error: Unexpected " + currentToken.Value) // Syntax error
		return ASTnode{}
	}

	node := ASTnode{symbol: nonTerminal}

	rules := strings.Split(rule, " ")
	for _, symbol := range rules[1:] { // Discard first element (current non-terminal)
		_, exists := nonTerminalMap[symbol]
		if exists {
			child := parse(symbol) // Recursively parse non-terminals
			if child.symbol == "" {
				return ASTnode{} // Return up call stack if error occured
			}
			node.addChild(child)
		} else if symbol == currentToken.Value {
			node.addChild(ASTnode{symbol: symbol})
			currentToken = retrieveAndValidateToken()
			if currentToken.Type == "ERROR" { // Invalid token, stop syntax parsing
				fmt.Println("Error: Invalid token", currentToken.Value)
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

func retrieveAndValidateToken() Token {
	token := lexer.GetNextToken()
	if token.Type == "IDENTIFIER" || token.Type == "INTEGER" || token.Type == "DOUBLE" {
		token.Value = token.Type // Syntax parsing only needs to know type classification
	}
	return token
}

func writeSyntaxError(err string) {
	errorMessage := fmt.Sprintf("%s at line %d column %d", err, lexer.LineNumber, lexer.ColumnNumber)
	fmt.Println(errorMessage)
	fmt.Fprintln(syntaxErrorFile, errorMessage)
}

func StartParser() {
	var err error
	syntaxErrorFile, err = os.Create("parser/syntax_errors.txt")
	if err != nil {
		log.Fatal("Error creating syntax_errors.txt")
	}
	currentToken = retrieveAndValidateToken() // get first token
	root := parse("program")
	if root.symbol == "" && !lexer.IsPanicMode {
		lexer.ErrorRecovery() // Init error recovery for lexer
		return
	}
}

func CloseParserFiles() {
	defer syntaxErrorFile.Close()
}
