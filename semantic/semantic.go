package semantic

import (
	"Compiler/parser"
	"fmt"
	"os"
	"strings" // Import strings package for indentation

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
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
	Parent  *SymbolTable // Pointer to parent scope (for nested scopes)
}

var currentType string
var inDeclContext bool
var currentFunctionReturnType string // Track current function's return type

// Add function to track function parameters
type FunctionInfo struct {
	ReturnType string
	Parameters map[string]string // parameter name -> type
}

var currentFunction *FunctionInfo

var scopeStack []*SymbolTable
var allScopes []*SymbolTable

func pushScope() {
	newScope := &SymbolTable{Entries: make(map[string]SymbolTableEntry), Parent: currentScope()}
	scopeStack = append(scopeStack, newScope)
	allScopes = append(allScopes, newScope)
}

func popScope() {
	scopeStack = scopeStack[:len(scopeStack)-1]
}

func currentScope() *SymbolTable {
	if len(scopeStack) == 0 {
		return nil
	}
	return scopeStack[len(scopeStack)-1]
}

func addSymbol(typ string, token string, lexeme string, params map[string]string) {
	current := currentScope()
	if current == nil {
		fmt.Println("Error: No active scope to add symbol")
		return
	}

	// Check for redeclaration in the current scope
	if _, exists := current.Entries[lexeme]; exists {
		fmt.Printf("Error: Redeclaration of variable '%s'", lexeme)
		return
	}

	// Add the symbol to the current scope
	current.Entries[lexeme] = SymbolTableEntry{
		Line:   1,
		Lexeme: lexeme,
		Token:  token,
		Type:   typ,
		Params: params,
	}
}

func retrieveIdentifier(identifier string) SymbolTableEntry {
	current := currentScope()
	if current == nil {
		fmt.Println("ERROR: No active scope when retrieving identifier", identifier)
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
		fmt.Println("ERROR: Undeclared variable", identifier)
		return SymbolTableEntry{}
	}
	return current.Entries[identifier]
}

// Gets the first type in the children
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

// Sets to double if any doubles or /, otherwise int
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

// Helper function to check type compatibility
func isTypeCompatible(type1 string, type2 string) bool {
	fmt.Println("type1", type1, "type2", type2)
	if type1 == type2 {
		return true
	}
	if type1 == "double" && type2 == "int" {
		return true
	}
	return false
}

func WalkAST(node *ASTnode) {
	// Initialize scope stack if empty
	if len(scopeStack) == 0 {
		pushScope()
	}

	if node == nil {
		return
	}

	// New scope for blocks (def, if, while)
	if node.Type == "KEYWORD" && (node.Lexeme == "if" || node.Lexeme == "else" || node.Lexeme == "while" || node.Lexeme == "def") {
		pushScope()
	}

	// Track function return type and parameters when entering a function definition
	if node.Symbol == "fdec" {
		// The return type is the first child after 'def'
		currentFunctionReturnType = node.Children[1].Children[0].Lexeme
		currentFunction = &FunctionInfo{
			ReturnType: currentFunctionReturnType,
			Parameters: make(map[string]string),
		}

		funcName := node.Children[2].Children[0].Lexeme

		paramsNode := node.Children[4]

		// Recursively process params
		var processParams func(*ASTnode)
		processParams = func(pNode *ASTnode) {
			if pNode == nil || len(pNode.Children) < 3 {
				return
			}

			// Extract type
			typeNode := pNode.Children[0]            // type
			paramType := typeNode.Children[0].Lexeme // int / double

			// Extract identifier
			idNode := pNode.Children[1].Children[0] // id
			paramName := idNode.Children[0].Lexeme  // IDENTIFIER

			// Add to symbol table and function param list
			currentFunction.Parameters[paramName] = paramType
			addSymbol(paramType, "IDENTIFIER", paramName, nil)

			// Recurse into params'
			if len(pNode.Children) > 2 {
				paramsPrime := pNode.Children[2]
				if len(paramsPrime.Children) > 1 {
					// The second child is the next `params`
					processParams(&paramsPrime.Children[1])
				}
			}
		}

		processParams(&paramsNode)
		addSymbol(currentFunctionReturnType, "IDENTIFIER", funcName, currentFunction.Parameters)

	}

	// Start of a declaration
	if node.Symbol == "decl" {
		inDeclContext = true
	}

	// Save the type for the declaration
	if node.Symbol == "type" && len(node.Children) > 0 {
		currentType = node.Children[0].Lexeme
	}

	// Handle identifiers
	if node.Symbol == "IDENTIFIER" {
		if inDeclContext {
			addSymbol(currentType, node.Symbol, node.Lexeme, nil) // declaration
			fmt.Printf("Declared %s as %s\n", node.Lexeme, currentType)
		} else {
			// Check if it's a function call
			if len(node.Children) > 0 && node.Children[0].Symbol == "(" { // will never happen
				// This is a function call, we'll handle it in the function call section
				entry := retrieveIdentifier(node.Lexeme)
				if entry.Type == "" {
					fmt.Printf("Error: Undeclared function '%s'\n", node.Lexeme)
				}
			}
		}
	}

	if node.Symbol == "factor" {
		if len(node.Children) <= 1 {
			return
		}

		id := node.Children[0]
		factorPrime := node.Children[1]

		// Check if it's a function call
		if id.Symbol == "id" && len(factorPrime.Children) > 0 && factorPrime.Children[0].Lexeme == "(" {
			funcName := id.Children[0].Lexeme // Extract IDENTIFIER from id
			funcEntry := retrieveIdentifier(funcName)
			if funcEntry.Type == "" {
				fmt.Printf("Error: Undeclared function '%s'\n", funcName)
				return
			}

			// Get declared parameter types
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
						collectExprTypes(&node.Children[1]) // exprseq'
					}
				} else if node.Symbol == "exprseq'" && len(node.Children) > 1 {
					collectExprTypes(&node.Children[1])
				}
			}

			collectExprTypes(&exprseq)

			// Check number of arguments
			if len(argTypes) != len(paramTypes) {
				fmt.Printf("Error: Function '%s' expects %d arguments but got %d\n", funcName, len(paramTypes), len(argTypes))
				return
			}

			// Check argument types
			for i := range argTypes {
				if !isTypeCompatible(paramTypes[i], argTypes[i]) {
					fmt.Printf("Type Error: Argument %d to function '%s' expects type '%s' but got '%s'\n", i+1, funcName, paramTypes[i], argTypes[i])
				}
			}
		}
	}

	// Handle assignments
	if node.Symbol == "statement" && len(node.Children) > 1 && node.Children[1].Symbol == "=" {
		varType := getVarType(&node.Children[0])
		exprType := getExprType(&node.Children[2])
		if !isTypeCompatible(varType, exprType) {
			fmt.Printf("Type Error: Cannot assign %s to variable of type %s\n", exprType, varType)
		}
	}

	// TODO: I don't think they could ever be invalid for comparison. Just implicitly promote ints to doubles
	// Handle comparisons
	// if node.Symbol == "bfactor" && len(node.Children) > 0 && node.Children[0].Symbol == "(" {
	// 	leftType := getExprType(&node.Children[1])
	// 	rightType := getExprType(&node.Children[3])
	// 	if !isTypeCompatible(leftType, rightType) {
	// 		fmt.Printf("Type Error: Cannot compare %s with %s\n", leftType, rightType)
	// 	}
	// }

	// Handle return statements
	if node.Symbol == "statement" && len(node.Children) > 0 && node.Children[0].Symbol == "return" {
		if currentFunctionReturnType == "" {
			fmt.Println("Error: Return statement outside of function")
			return
		}
		returnType := getExprType(&node.Children[1])
		if !isTypeCompatible(currentFunctionReturnType, returnType) {
			fmt.Printf("Type Error: Function declared to return %s but returning %s\n",
				currentFunctionReturnType, returnType)
		}
	}

	// Visit children
	for i := range node.Children {
		WalkAST(&node.Children[i])
	}

	// End of decl subtree
	if node.Symbol == "decl" {
		inDeclContext = false
		currentType = ""
	}

	// Reset function return type when exiting function definition
	if node.Type == "KEYWORD" && node.Lexeme == "fed" {
		currentFunctionReturnType = ""
		currentFunction = nil
		popScope() // Exit the current scope
	}

	if node.Type == "KEYWORD" && (node.Lexeme == "fi" || node.Lexeme == "od") {
		popScope() // Exit the current scope
	}
}

// VisualizeAST prints a text-based representation of the AST with improved formatting
func VisualizeAST(root *ASTnode) {
	fmt.Println("\n=== Abstract Syntax Tree Visualization ===\n")
	visualizeNode(root, "", true)
	fmt.Println("\n=========================================\n")
}

// visualizeNode is a recursive helper function for VisualizeAST
func visualizeNode(node *ASTnode, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// Create the tree-like structure
	marker := "└── "
	if !isLast {
		marker = "├── "
	}

	// Print the current node with appropriate formatting
	nodeInfo := fmt.Sprintf("%s%s%s", prefix, marker, node.Symbol)

	// Add additional node information with colors
	details := []string{}
	if node.Lexeme != "" {
		details = append(details, color.GreenString("lexeme=%s", node.Lexeme))
	}
	if node.Type != "" {
		details = append(details, color.BlueString("type=%s", node.Type))
	}

	if len(details) > 0 {
		nodeInfo += " (" + strings.Join(details, ", ") + ")"
	}

	fmt.Println(nodeInfo)

	// Calculate the prefix for child nodes
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// Recursively visualize children
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		visualizeNode(&child, childPrefix, isLastChild)
	}
}

// PrintSymbolTable prints the current symbol table in a formatted way
func PrintSymbolTable() {
	if len(allScopes) == 0 {
		fmt.Println("No symbol tables available")
		return
	}

	fmt.Println("\n=== Symbol Table Hierarchy ===\n")

	// Track visited scopes to prevent infinite loops
	visited := make(map[*SymbolTable]bool)

	var printScope func(scope *SymbolTable, level int)
	printScope = func(scope *SymbolTable, level int) {
		if scope == nil || visited[scope] {
			return
		}
		visited[scope] = true

		indent := strings.Repeat("  ", level)
		fmt.Printf("%sScope Level %d:\n", indent, level)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Identifier", "Type", "Token", "Line", "Params"})

		for name, entry := range scope.Entries {
			table.Append([]string{
				color.GreenString(name),
				color.BlueString(entry.Type),
				entry.Token,
				fmt.Sprintf("%d", entry.Line),
				fmt.Sprintf("%v", entry.Params),
			})
		}

		table.SetAutoFormatHeaders(false)
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnSeparator("|")
		table.SetHeaderLine(false)

		table.Render()
		fmt.Println()

		// Recursively print child scopes
		for _, child := range allScopes {
			if child.Parent == scope {
				printScope(child, level+1)
			}
		}
	}

	// Print all top-level scopes
	for _, scope := range allScopes {
		if scope.Parent == nil {
			printScope(scope, 0)
		}
	}
}

func SemanticAnalysis(root *ASTnode) {
	WalkAST(root)
	PrintSymbolTable()
}
