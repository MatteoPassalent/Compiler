package semantic

import (
	"Compiler/parser"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	semanticErrorFile *os.File
)

type ASTnode = parser.ASTnode

type SymbolTableEntry struct {
	Line   int
	Lexeme string
	Token  string
	Type   string
	Params map[string]string
}

type SymbolTable struct {
	Entries map[string]SymbolTableEntry
	Parent  *SymbolTable
}

var currentType string
var inDeclContext bool
var currentFunctionReturnType string

type FunctionInfo struct {
	ReturnType string
	Parameters map[string]string
}

var currentFunction *FunctionInfo

var scopeStack []*SymbolTable
var allScopes []*SymbolTable

/**
 * Function to push a new scope onto the stack
 * This allows for nested scopes in the symbol table
 * Each new scope can have its own set of symbols
 */
func pushScope() {
	newScope := &SymbolTable{Entries: make(map[string]SymbolTableEntry), Parent: currentScope()}
	scopeStack = append(scopeStack, newScope)
	allScopes = append(allScopes, newScope)
}

/**
 * Function to pop the current scope off the stack
 * This is used when exiting a block of code
 * The current scope is removed from the stack
 * and the parent scope becomes the current scope
 */
func popScope() {
	scopeStack = scopeStack[:len(scopeStack)-1]
}

/**
 * Function to get the current scope
 * This returns the top scope on the stack
 * If the stack is empty, it returns nil
 */
func currentScope() *SymbolTable {
	if len(scopeStack) == 0 {
		return nil
	}
	return scopeStack[len(scopeStack)-1]
}

/**
 * Function to add a symbol to the current scope
 * Checks for redeclaration and adds the symbol to the current scope
 */
func addSymbol(typ string, token string, lexeme string, params map[string]string, line int) {
	current := currentScope()
	if current == nil {
		fmt.Fprintln(semanticErrorFile, "Error: No active scope to add symbol")
		return
	}

	// Check for redeclaration in the current scope
	if _, exists := current.Entries[lexeme]; exists && token == "IDENTIFIER" {
		fmt.Fprintf(semanticErrorFile, "Error: Redeclaration of variable '%s'\n", lexeme)
		return
	}

	// Add the symbol to the current scope
	current.Entries[lexeme] = SymbolTableEntry{
		Line:   line,
		Lexeme: lexeme,
		Token:  token,
		Type:   typ,
		Params: params,
	}
}

/**
 * Function to retrieve a symbol from the current scope
 * Checks the current scope and its parent scopes
 * If the symbol is not found, it returns an empty SymbolTableEntry
 */
func retrieveIdentifier(identifier string) SymbolTableEntry {
	current := currentScope()
	if current == nil {
		fmt.Fprintln(semanticErrorFile, "ERROR: No active scope when retrieving identifier", identifier)
		return SymbolTableEntry{}
	}

	_, found := current.Entries[identifier]
	for !found && current != nil {
		current = current.Parent
		if current != nil {
			_, found = current.Entries[identifier]
		}
	}
	if !found {
		fmt.Fprintf(semanticErrorFile, "ERROR: Undeclared variable '%s'\n", identifier)
		return SymbolTableEntry{}
	}
	return current.Entries[identifier]
}

/**
 * Traverses children to find the type of a variable
 */
func getVarType(node *ASTnode) string {
	if node == nil {
		return ""
	}
	for _, child := range node.Children {
		typ := getVarType(&child)
		if typ != "" {
			return typ
		}
	}

	// Handle literals
	if node.Type == "INTEGER" {
		return "int"
	}
	if node.Type == "DOUBLE" {
		return "double"
	}
	// Handle identifiers
	if node.Symbol == "IDENTIFIER" {
		entry := retrieveIdentifier(node.Lexeme)
		return entry.Type
	}
	return ""

}

/**
 * Retrieves the type of an expression
 * Upgrades type to double if expression contains a double
 * or if the expression includes a division operator
 */
func getExprType(node *ASTnode) string {
	if node == nil {
		return ""
	}

	for _, child := range node.Children {
		typ := getExprType(&child)
		if typ == "double" {
			return "double"
		}
	}
	if node.Type == "DOUBLE" {
		return "double"
	}
	if node.Symbol == "/" {
		return "double"
	}
	if node.Symbol == "IDENTIFIER" {
		entry := retrieveIdentifier(node.Lexeme)
		return entry.Type
	}

	return "int"
}

/**
 * Checks if two types are compatible
 * Allows double to be assigned to int
 */
func isTypeCompatible(type1 string, type2 string) bool {
	if type1 == type2 {
		return true
	}
	if type1 == "double" && type2 == "int" {
		return true
	}
	return false
}

/**
 * Builds the symbol tables for the AST, prints tables to file
 * Logs errors to file and continues in recovery
 * Checks for:
 * 	- Redeclarations errors
 * 	- Undeclared variable reference errors
 * 	- Variable assignment type errors
 * 	- Function arguments type errors
 * 	- Function return type errors
 * 	- Upgrades integers to doubles during comparison
 * 	  operation so all comparisons are type valid
 */
func BuildSymbolTables(node *ASTnode) {
	// Initialize scope stack if empty
	if len(scopeStack) == 0 {
		pushScope()
	}

	// Base case
	if node == nil {
		return
	}

	// Handle scopes and keywords
	if node.Type == "KEYWORD" {
		addSymbol("keyword", "KEYWORD", node.Lexeme, nil, node.Line)

		// New scope for if, else, and while blocks
		if node.Lexeme == "if" || node.Lexeme == "while" {
			pushScope()
		}
		if node.Lexeme == "else" {
			popScope()  // close the then block
			pushScope() // open a fresh scope for else
		}
		if node.Lexeme == "fi" || node.Lexeme == "od" {
			popScope() // close the if or while block
		}
		// Reset current function, current return type, and pop scope when exiting function body
		if node.Lexeme == "fed" {
			currentFunctionReturnType = ""
			currentFunction = nil
			popScope()
		}
	}

	// Add literals to symbol table
	if node.Type == "INTEGER" || node.Type == "DOUBLE" {
		typ := "int"
		if node.Type == "DOUBLE" {
			typ = "double"
		}
		addSymbol(typ,
			node.Type,
			node.Lexeme, nil, node.Line)
	}

	// Updated to ensure function declarations and parameters are in separate scopes
	if node.Symbol == "fdec" {
		currentFunctionReturnType = node.Children[1].Children[0].Lexeme
		currentFunction = &FunctionInfo{
			ReturnType: currentFunctionReturnType,
			Parameters: make(map[string]string),
		}

		funcName := node.Children[2].Children[0].Lexeme

		// Add function declaration to the current (top-level) scope
		addSymbol(currentFunctionReturnType, "IDENTIFIER", funcName, currentFunction.Parameters, node.Line)

		// Create a new scope for the function body
		pushScope()

		paramsNode := node.Children[4]

		// Recursively process parameters
		var processParams func(*ASTnode)
		processParams = func(pNode *ASTnode) {
			// Base case
			if pNode == nil || len(pNode.Children) < 3 {
				return
			}

			// Extract type
			typeNode := pNode.Children[0]
			paramType := typeNode.Children[0].Lexeme

			// Extract identifier
			idNode := pNode.Children[1].Children[0]
			paramName := idNode.Children[0].Lexeme

			// Add parameter to the function's scope
			currentFunction.Parameters[paramName] = paramType
			addSymbol(paramType, "IDENTIFIER", paramName, nil, node.Line)

			// Recurse into params prime
			if len(pNode.Children) > 2 {
				paramsPrime := pNode.Children[2]
				if len(paramsPrime.Children) > 1 {
					processParams(&paramsPrime.Children[1])
				}
			}
		}
		processParams(&paramsNode)
	}

	// Start of a declaration
	if node.Symbol == "decl" {
		inDeclContext = true
	}

	// Save the type for the declaration
	if node.Symbol == "type" && len(node.Children) > 0 {
		currentType = node.Children[0].Lexeme
	}

	// Handle identifier declarations if in declaration context
	if node.Symbol == "IDENTIFIER" {
		if inDeclContext {
			addSymbol(currentType, node.Symbol, node.Lexeme, nil, node.Line)
		} else {
			// Check if variable is declared
			retrieveIdentifier(node.Lexeme)
		}
	}

	if node.Symbol == "factor" {
		if len(node.Children) == 1 {
			BuildSymbolTables(&node.Children[0])
			return
		}

		id := node.Children[0]
		factorPrime := node.Children[1]

		// Check for function call
		if id.Symbol == "id" && len(factorPrime.Children) > 0 && factorPrime.Children[0].Lexeme == "(" {
			funcName := id.Children[0].Lexeme
			funcEntry := retrieveIdentifier(funcName)
			if funcEntry.Type == "" {
				fmt.Fprintf(semanticErrorFile, "Error: Undeclared function '%s'\n", funcName)
				return
			}

			// Retrieve parameter types from function entry
			paramTypes := []string{}
			if funcEntry.Params != nil {
				for _, typ := range funcEntry.Params {
					paramTypes = append(paramTypes, typ)
				}
			}

			// Extract argument types from exprseq
			exprseq := factorPrime.Children[1]
			argTypes := []string{}

			var collectExprTypes func(*ASTnode)
			collectExprTypes = func(node *ASTnode) {
				if node.Symbol == "exprseq" && len(node.Children) > 0 {
					argTypes = append(argTypes, getExprType(&node.Children[0]))
					if len(node.Children) > 1 {
						collectExprTypes(&node.Children[1])
					}
				} else if node.Symbol == "exprseq'" && len(node.Children) > 1 {
					collectExprTypes(&node.Children[1])
				}
			}
			collectExprTypes(&exprseq)

			// Check for correct number of arguments
			if len(argTypes) != len(paramTypes) {
				fmt.Fprintf(semanticErrorFile, "Error: Function '%s' expects %d arguments but got %d\n", funcName, len(paramTypes), len(argTypes))
				return
			}

			// Validate argument types match parameter types
			for i := range argTypes {
				if !isTypeCompatible(paramTypes[i], argTypes[i]) {
					fmt.Fprintf(semanticErrorFile, "Type Error: Argument %d to function '%s' expects type '%s' but got '%s'\n", i+1, funcName, paramTypes[i], argTypes[i])
				}
			}
		}
	}

	// Handle variable assignments
	if node.Symbol == "statement" && len(node.Children) > 1 && node.Children[1].Symbol == "=" {
		varType := getVarType(&node.Children[0])
		exprType := getExprType(&node.Children[2])
		// Check compatability between LHS and RHS
		if !isTypeCompatible(varType, exprType) {
			fmt.Fprintf(semanticErrorFile, "Type Error: Cannot assign %s to variable of type %s\n", exprType, varType)
		}
	}

	// Handle return statements
	if node.Symbol == "statement" && len(node.Children) > 0 && node.Children[0].Symbol == "return" {
		if currentFunctionReturnType == "" {
			fmt.Fprintln(semanticErrorFile, "Error: Return statement outside of function")
			return
		}
		returnType := getExprType(&node.Children[1])
		// Check compatibility between return type and function return type
		if !isTypeCompatible(currentFunctionReturnType, returnType) {
			fmt.Fprintf(semanticErrorFile, "Type Error: Function declared to return %s but returning %s\n",
				currentFunctionReturnType, returnType)
		}
	}

	// Type check for boolean comparisons
	if node.Symbol == "bfactor" && len(node.Children) == 5 && node.Children[2].Symbol == "comp" {
		leftType := getExprType(&node.Children[1])
		rightType := getExprType(&node.Children[3])

		if leftType != rightType || (leftType != "int" && leftType != "double") {
			fmt.Fprintf(semanticErrorFile, "Type Error: Invalid comparison between '%s' and '%s'\n", leftType, rightType)
		}
	}

	// Recurse AST
	for i := range node.Children {
		BuildSymbolTables(&node.Children[i])
	}

	// End of declaration subtree
	if node.Symbol == "decl" {
		inDeclContext = false
		currentType = ""
	}

}

/**
 * Visualizes the symbol table in a text file
 */
func PrintSymbolTable() {
	if len(allScopes) == 0 {
		fmt.Println("No symbol tables available")
		return
	}

	file, err := os.Create("semantic/symbol_tables.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	visited := make(map[*SymbolTable]bool)

	formatParams := func(params map[string]string) string {
		if len(params) == 0 {
			return "[]"
		}
		result := "["
		for name, typ := range params {
			result += fmt.Sprintf("%s: %s, ", name, typ)
		}
		result = strings.TrimSuffix(result, ", ") + "]"
		return result
	}

	var printScope func(scope *SymbolTable, level int)
	printScope = func(scope *SymbolTable, level int) {
		if scope == nil || visited[scope] {
			return
		}
		visited[scope] = true

		indent := strings.Repeat("  ", level)
		fmt.Fprintf(file, "%s+-----------------------------+----------------+------------+------+------------------------------+\n", indent)
		fmt.Fprintf(file, "%s| Scope Level %d               |\n", indent, level)
		fmt.Fprintf(file, "%s+-----------------------------+----------------+------------+------+------------------------------+\n", indent)
		fmt.Fprintf(file, "%s| Lexeme                      | Type           | Token      | Line | Params                       |\n", indent)
		fmt.Fprintf(file, "%s+-----------------------------+----------------+------------+------+------------------------------+\n", indent)

		for name, entry := range scope.Entries {
			fmt.Fprintf(file, "%s| %-27s | %-14s | %-10s | %-4d | %-28s |\n",
				indent, name, entry.Type, entry.Token, entry.Line, formatParams(entry.Params))
		}
		fmt.Fprintf(file, "%s+-----------------------------+----------------+------------+------+------------------------------+\n\n", indent)

		for _, child := range allScopes {
			if child.Parent == scope {
				printScope(child, level+1)
			}
		}
	}

	for _, scope := range allScopes {
		if scope.Parent == nil {
			printScope(scope, 0)
		}
	}
}

func SemanticAnalysis(root *ASTnode) {
	var err error

	semanticErrorFile, err = os.Create("semantic/semantic_errors.txt")
	if err != nil {
		log.Fatal("Error creating error file:", err)
	}
	BuildSymbolTables(root)
	PrintSymbolTable()
	defer semanticErrorFile.Close()
}
