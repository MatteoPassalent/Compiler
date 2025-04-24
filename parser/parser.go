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
	Line     int
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
		writeSyntaxError("Error: Unexpected " + currentToken.Lexeme) // Syntax error
		currentToken = getNextValidToken()
		currentTokenIndex, tokenExists = terminalMap[currentToken.Symbol]
		if !tokenExists {
			writeSyntaxError("Error: Invalid terminal " + currentToken.Lexeme) // Issue with LL1 or lexical
			return ASTnode{}
		}
		rule = ll1[nonTerminalIndex][currentTokenIndex]
		if rule == "" {
			writeSyntaxError("Error: Unexpected " + currentToken.Lexeme) // Syntax error
			return ASTnode{}
		}
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
			leafNode := ASTnode{Symbol: symbol, Lexeme: currentToken.Lexeme, Type: currentToken.Type, Line: currentToken.Line}
			node.addChild(leafNode)
			currentToken = getNextValidToken()
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

func getNextValidToken() Token {
	var validToken Token
	for {
		validToken = lexer.GetNextToken()
		if validToken.Type != "ERROR" {
			break
		}
	}
	return validToken
}

func writeSyntaxError(err string) {
	errorMessage := fmt.Sprintf("%s at line %d column %d", err, lexer.LineNumber, lexer.ColumnNumber)
	fmt.Fprintln(syntaxErrorFile, errorMessage)
}

func CloseParserFiles() {
	defer syntaxErrorFile.Close()
}

// Writes a text-based representation of the AST to a file
func VisualizeAST(root *ASTnode, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	visualizeNode(root, "", true, file)
}

// Recursive helper function for VisualizeAST
func visualizeNode(node *ASTnode, prefix string, isLast bool, file *os.File) {
	if node == nil {
		return
	}

	marker := "└── "
	if !isLast {
		marker = "├── "
	}

	nodeInfo := fmt.Sprintf("%s%s%s", prefix, marker, node.Symbol)

	details := []string{}
	if node.Lexeme != "" {
		details = append(details, fmt.Sprintf("lexeme=%s", node.Lexeme))
	}
	if node.Type != "" {
		details = append(details, fmt.Sprintf("type=%s", node.Type))
	}

	if len(details) > 0 {
		nodeInfo += " (" + strings.Join(details, ", ") + ")"
	}

	fmt.Fprintln(file, nodeInfo)

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		visualizeNode(&child, childPrefix, isLastChild, file)
	}
}

func StartParser() ASTnode {
	var err error
	syntaxErrorFile, err = os.Create("parser/syntax_errors.txt")
	if err != nil {
		log.Fatal("Error creating syntax_errors.txt")
	}
	currentToken = getNextValidToken()
	root := parse("program")
	if root.Symbol == "" && !lexer.IsPanicMode {
		lexer.ErrorRecovery() // Init error recovery for lexer
		return ASTnode{}
	}
	VisualizeAST(&root, "parser/ast_visualization.txt")
	return root
}
