package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var currentToken Token

// var tokenPointer = 0

// var tokens = ReadTokens()

type Token struct {
	Type  string
	value string
}

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

// func getNextToken() Token {
// 	if tokenPointer >= len(tokens) {
// 		return Token{Type: "EOF", value: "EOF"}
// 	}
// 	token := tokens[tokenPointer]
// 	if token.Type == "IDENTIFIER" || token.Type == "INTEGER" || token.Type == "DOUBLE" {
// 		token.value = token.Type
// 	}
// 	tokenPointer++
// 	return token
// }

func parse(nonTerminal string) ASTnode {
	nonTerminalIndex, exists := nonTerminalMap[nonTerminal]
	if !exists {
		fmt.Println("Error: Invalid non-terminal", nonTerminal)
		return ASTnode{}
	}
	currentTokenIndex, tokenExists := terminalMap[currentToken.value]
	if !tokenExists {
		fmt.Println("Error: Invalid terminal", currentToken)
		return ASTnode{}
	}
	rule := ll1[nonTerminalIndex][currentTokenIndex]
	if rule == "" {
		fmt.Println("Error: Invalid rule")
		return ASTnode{}
	}

	node := ASTnode{symbol: nonTerminal}

	rules := strings.Split(rule, " ")
	for _, symbol := range rules[1:] {
		_, exists := nonTerminalMap[symbol]
		if exists {
			child := parse(symbol)
			if child.symbol == "" {
				fmt.Print("Error: Invalid child")
				return ASTnode{}
			}
			node.addChild(child)
		} else if symbol == currentToken.value {
			node.addChild(ASTnode{symbol: symbol})
			currentToken = retrieveAndValidateToken()
			if currentToken.Type == "EOF" {
				if nonTerminal == "program" && symbol == "." {
					break
				} else {
					fmt.Println("Error: Unexpected EOF")
					return ASTnode{}
				}
			}
		} else if symbol == "e" {
			continue
		} else {
			fmt.Println("Error: Invalid symbol", symbol)
			return ASTnode{}
		}
	}
	return node
}

// // ReadTokens reads a file and returns a slice of Token structs
// func ReadTokens() []Token {
// 	file, _ := os.Open("../Lexical_analysis/output.txt")

// 	defer file.Close()

// 	var tokens []Token
// 	scanner := bufio.NewScanner(file)

// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		parts := strings.Fields(line) // Split by whitespace

// 		if len(parts) < 4 {
// 			continue // Skip malformed lines
// 		}

// 		tokenType := parts[1]  // Second word is Type
// 		tokenValue := parts[3] // Fourth word is Token

// 		tokens = append(tokens, Token{Type: tokenType, value: tokenValue})
// 	}

// 	return tokens
// }

func retrieveAndValidateToken() Token {
	token := getNextToken()
	if token.Type == "ERROR" {
		fmt.Println("Error: Invalid token")
		return Token{}
	}
	if token.Type == "IDENTIFIER" || token.Type == "INTEGER" || token.Type == "DOUBLE" {
		token.value = token.Type
	}
	fmt.Println(token)
	return token
}

func main() {
	// Open input file
	var err error
	file, err = os.Open("../Testing/Test1.cp")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()
	initializeInputTypeTable()
	fillBuffer()
	currentToken = retrieveAndValidateToken()
	fmt.Println(parse("program"))
}
