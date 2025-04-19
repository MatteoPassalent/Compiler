package codegen

import (
	"Compiler/parser"
	"fmt"
	"os"
)

type ASTnode = parser.ASTnode

type TACInstruction struct {
	Op   string // "ADD", "SUB", "MUL", "CALL", etc
	Arg1 string
	Arg2 string
	Arg3 string
}

var tempCount int
var labelCount int
var file *os.File

var currentFunctionName string

// Creates new temporary variable names
func MakeTemp() string {
	tempCount++
	return fmt.Sprintf("t%d", tempCount)
}

// Creates new label names
func MakeLabel() string {
	labelCount++
	return fmt.Sprintf("L%d", labelCount)
}

// Appends new 3TAC instructions
func Emit(op, arg1, arg2, arg3 string) {
	fmt.Fprintln(file, op, arg1, arg2, arg3)
}

// Generate 3TAC code
func GenerateTAC(astRoot *ASTnode) {
	var err error
	file, err = os.Create("codegen/IR_code.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	tempCount = 0
	labelCount = 0
	genProgram(astRoot)
	defer file.Close()
}

// Handles functions and main program statements
func genProgram(node *ASTnode) {
	Emit("B", "main", "", "")
	for _, child := range node.Children {
		switch child.Symbol {
		case "fdecls", "fdecls'":
			genFdecls(&child)
		case "statement_seq":
			Emit("main:", "", "", "")
			Emit("Begin", "", "", "")
			genStatementSeq(&child)
		}
	}
}

// Handles function declarations
func genFdecls(node *ASTnode) {
	for _, child := range node.Children {
		if child.Symbol == "fdec" {
			genFdec(&child)
		} else if child.Symbol == "fdecls'" {
			genFdecls(&child)
		}
	}
}

// Handles a single function declaration
func genFdec(node *ASTnode) {

	funcNameNode := node.Children[2] //  "id"
	funcName := funcNameNode.Children[0].Lexeme

	currentFunctionName = funcName

	// function start
	Emit(fmt.Sprintf("%s:", funcName), "", "", "")
	Emit("Begin", "", "", "")
	Emit("Push", "{LR}", "", "") // Push the link register
	Emit("Push", "{FP}", "", "") // Push the frame pointer

	fpCounter := 4

	paramsNode := node.Children[4]

	var processParams func(*ASTnode)
	processParams = func(pNode *ASTnode) {
		if pNode == nil || len(pNode.Children) < 3 {
			return
		}

		idNode := pNode.Children[1].Children[0] // id
		paramName := idNode.Children[0].Lexeme  // IDENTIFIER

		fpCounter += 4
		Emit("ADD", paramName, "{FP}", fmt.Sprintf("%d", fpCounter)) // Retrieve parameter

		// Recurse into params'
		if len(pNode.Children) > 2 {
			paramsPrime := pNode.Children[2]
			if len(paramsPrime.Children) > 1 {
				processParams(&paramsPrime.Children[1])
			}
		}
	}
	processParams(&paramsNode)

	// Handle function body
	for _, child := range node.Children {
		switch child.Symbol {
		case "statement_seq":
			genStatementSeq(&child)
		}
	}
	Emit(funcName+"Exit:", "", "", "")
	Emit("POP", "{FP}", "", "")
	Emit("POP", "{PC}", "", "")
	currentFunctionName = ""
}

// Handles sequence of statements
func genStatementSeq(node *ASTnode) {
	if node == nil || node.Symbol != "statement_seq" {
		return
	}
	if len(node.Children) == 0 {
		return
	}

	// Generate instructions for first statement
	statementNode := &node.Children[0]
	genStatement(statementNode)

	// Handle subsequent statements
	if len(node.Children) > 1 {
		statementSeqPrime := &node.Children[1]
		genStatementSeqPrime(statementSeqPrime)
	}
}

// Handles additional statements in a sequence
func genStatementSeqPrime(node *ASTnode) {
	if node == nil || node.Symbol != "statement_seq'" {
		return
	}
	if len(node.Children) == 0 {
		return
	}

	semicolonNode := &node.Children[0]
	if semicolonNode.Lexeme == ";" {
		statementSeqNode := &node.Children[1]
		genStatementSeq(statementSeqNode)
	}
}

// Handles a signle statement
func genStatement(node *ASTnode) {
	if len(node.Children) == 0 {
		return
	}
	// Handle assignment statement
	if len(node.Children) >= 3 && node.Children[1].Symbol == "=" {
		varNode := &node.Children[0]
		exprNode := &node.Children[2]

		// Generate instructions for the expression
		exprResult := genExpr(exprNode)

		// Get the variable name
		varName := getVarName(varNode)

		// Final assignment instruction
		Emit("MOV", exprResult, varName, "")
		return
	}

	// Handle if statement
	if node.Children[0].Symbol == "if" {
		bexprNode := &node.Children[1]
		trueSeqNode := &node.Children[3]

		elseOrFiNode := &node.Children[4]

		labelTrue := MakeLabel()
		labelFalse := MakeLabel()

		// Generate instructions for bool expression
		condResult := genBexpr(bexprNode)

		// Add conditional branch jump instruction
		Emit("B"+condResult, labelTrue, "", "")
		// Branch to false block if condition is false
		Emit("B", labelFalse, "", "")

		// True block
		Emit(labelTrue+":", "", "", "")
		genStatementSeq(trueSeqNode)

		// False block
		Emit(labelFalse+":", "", "", "")
		genElsePart(elseOrFiNode)
		return
	}

	// Handle while statement
	if node.Children[0].Symbol == "while" {
		condNode := &node.Children[1]
		bodyNode := &node.Children[3]

		loopLabel := MakeLabel()
		bodyLabel := MakeLabel()
		exitLabel := MakeLabel()

		// Loop start
		Emit(loopLabel+":", "", "", "")
		// Generate instructions for the condition
		condResult := genBexpr(condNode)

		Emit("B"+condResult, bodyLabel, "", "") // Jump to body if true
		Emit("B", exitLabel, "", "")            // Jump to exit if false

		// While body
		Emit(bodyLabel+":", "", "", "")
		genStatementSeq(bodyNode)
		Emit("B", loopLabel, "", "") // Jump back to loop start

		Emit(exitLabel+":", "", "", "") // Exit label

		return
	}

	// Handle print statement
	if node.Children[0].Symbol == "print" {
		exprNode := &node.Children[1]
		exprResult := genExpr(exprNode)
		Emit("PRINT", exprResult, "", "")
		return
	}

	// Handle return statement
	if node.Children[0].Symbol == "return" {
		exprNode := &node.Children[1]
		exprResult := genExpr(exprNode)
		Emit("MOV", "{fp - 4}", exprResult, "")       // Store return value
		Emit("B", currentFunctionName+"Exit", "", "") // Jump to function exit
		return
	}
}

// Generates statement instructions for else part of an if statement OR fi
func genElsePart(node *ASTnode) {
	if len(node.Children) > 0 && node.Children[0].Symbol == "else" {
		seqNode := &node.Children[1]
		genStatementSeq(seqNode)
	}
}

// Generates instructions for boolean expressions
func genBexpr(node *ASTnode) string {
	if node == nil || node.Symbol != "bexpr" || len(node.Children) == 0 {
		return ""
	}

	left := genBterm(&node.Children[0])

	if len(node.Children) == 1 {
		return left
	}
	return genBexprPrime(&node.Children[1], left)
}

func genBexprPrime(node *ASTnode, leftTmp string) string {
	if node == nil || len(node.Children) == 0 {
		return leftTmp // ε
	}

	rightTmp := genBterm(&node.Children[1])

	out := MakeTemp()
	Emit("OR", leftTmp, rightTmp, out)

	if len(node.Children) > 2 {
		return genBexprPrime(&node.Children[2], out)
	}
	return out
}

// Generates instructions for boolean terms
func genBterm(node *ASTnode) string {
	if node == nil || node.Symbol != "bterm" || len(node.Children) == 0 {
		return ""
	}
	left := genBfactor(&node.Children[0])

	if len(node.Children) == 1 {
		return left
	}
	return genBtermPrime(&node.Children[1], left)
}

func genBtermPrime(node *ASTnode, leftTmp string) string {
	if node == nil || len(node.Children) == 0 {
		return leftTmp
	}

	rightTmp := genBfactor(&node.Children[1])

	out := MakeTemp()
	Emit("AND", leftTmp, rightTmp, out)

	if len(node.Children) > 2 {
		return genBtermPrime(&node.Children[2], out)
	}
	return out
}

// Generates instructions for boolean factors
func genBfactor(node *ASTnode) string {
	if node == nil || node.Symbol != "bfactor" {
		return ""
	}

	if node.Children[0].Symbol == "not" {
		sub := genBfactor(&node.Children[1])
		out := MakeTemp()
		Emit("NOT", sub, "", out) // out = NOT sub
		return out
	}

	if len(node.Children) >= 5 && node.Children[0].Lexeme == "(" {
		left := genExpr(&node.Children[1])
		opTmp := node.Children[2]
		right := genExpr(&node.Children[3])

		// Extract actual comparison operator symbol:
		compTok := opTmp.Children[0].Lexeme
		if compTok == "<" {
			compTok = "LT"
		} else if compTok == ">" {
			compTok = "GT"
		} else if compTok == "<=" {
			compTok = "LE"
		} else if compTok == ">=" {
			compTok = "GE"
		} else if compTok == "==" {
			compTok = "EQ"
		} else if compTok == "<>" {
			compTok = "NE"
		}

		Emit("CMP", left, right, "")
		return compTok
	}
	return ""
}

// Generates instructions for expressions
func genExpr(node *ASTnode) string {
	if len(node.Children) == 2 && node.Symbol == "expr" {
		leftTemp := genTerm(&node.Children[0])
		return genExprPrime(&node.Children[1], leftTemp)
	}

	if len(node.Children) == 1 {
		return genExpr(&node.Children[0])
	}
	return ""
}

// Generate add and subtraction instructions
func genExprPrime(node *ASTnode, leftTemp string) string {
	if len(node.Children) == 0 {
		return leftTemp
	}

	op := node.Children[0].Symbol
	termNode := &node.Children[1]
	rightTemp := genTerm(termNode)

	newTemp := MakeTemp()
	switch op {
	case "+":
		Emit("ADD", newTemp, leftTemp, rightTemp)
	case "-":
		Emit("SUB", newTemp, leftTemp, rightTemp)
	}

	if len(node.Children) > 2 {
		return genExprPrime(&node.Children[2], newTemp)
	}

	return newTemp
}

func genTerm(node *ASTnode) string {
	if len(node.Children) == 2 && node.Symbol == "term" {
		leftTemp := genFactor(&node.Children[0])
		return genTermPrime(&node.Children[1], leftTemp)
	}
	if len(node.Children) == 1 {
		return genTerm(&node.Children[0])
	}
	return ""
}

// Generate multiplication, division, and modulus instructions
func genTermPrime(node *ASTnode, leftTemp string) string {
	if len(node.Children) == 0 {
		return leftTemp
	}

	op := node.Children[0].Symbol
	factorNode := &node.Children[1]
	rightTemp := genFactor(factorNode)

	newTemp := MakeTemp()
	switch op {
	case "*":
		Emit("MUL", newTemp, leftTemp, rightTemp)
	case "/":
		Emit("DIV", newTemp, leftTemp, rightTemp)
	case "%":
		Emit("MOD", newTemp, leftTemp, rightTemp)
	}

	if len(node.Children) > 2 {
		return genTermPrime(&node.Children[2], newTemp)
	}
	return newTemp
}

// Generates instructions for factors and returns the resulting variable
func genFactor(node *ASTnode) string {
	// Check for function calls
	if isCall, idN, exprSeqN := detectCall(node); isCall {
		return genFunctionCall(*idN, exprSeqN)
	}
	// Generate instruction to store integer literals in a temp
	if len(node.Children) == 1 && node.Children[0].Symbol == "INTEGER" {
		val := node.Children[0].Lexeme
		tmp := MakeTemp()
		Emit("MOV", val, tmp, "")
		return tmp
	}

	// Generate instruction to store double literals in a temp
	if len(node.Children) == 1 && node.Children[0].Symbol == "DOUBLE" {
		val := node.Children[0].Lexeme
		tmp := MakeTemp()
		Emit("MOV", val, tmp, "")
		return tmp
	}

	if node.Children[0].Symbol == "id" {
		return node.Children[0].Children[0].Lexeme
	}

	// handle ( expr )
	if len(node.Children) == 3 && node.Children[0].Lexeme == "(" {
		return genExpr(&node.Children[1])
	}

	if len(node.Children) == 1 {
		return genFactor(&node.Children[0])
	}
	return ""
}

// Detects if a factor is a function call
func detectCall(factor *ASTnode) (bool, *ASTnode, *ASTnode) {
	if factor == nil || factor.Symbol != "factor" || len(factor.Children) < 2 {
		return false, nil, nil
	}
	idNode := &factor.Children[0]
	factorPrime := &factor.Children[1]
	if len(factorPrime.Children) > 0 && factorPrime.Children[0].Lexeme == "(" {
		exprSeq := &factorPrime.Children[1]
		return true, idNode, exprSeq
	}
	return false, nil, nil
}

// Creates instructions for function calls
func genFunctionCall(idNode ASTnode, exprSeqNode *ASTnode) string {
	funcName := idNode.Children[0].Lexeme
	argTemps := collectArgs(exprSeqNode)

	// Push the arguments before the call
	for _, arg := range argTemps {
		Emit("PUSH", arg, "", "")
	}

	// Generate branch link instruction for the function call
	returnTemp := MakeTemp()
	Emit("BL", funcName, returnTemp, "")

	// Pop the arguments after the call
	for i := len(argTemps) - 1; i >= 0; i-- {
		arg := argTemps[i]
		Emit("POP", arg, "", "")
	}

	return returnTemp
}

// Returns the arguments from function calls
func collectArgs(exprSeqNode *ASTnode) []string {
	var result []string

	var walk func(*ASTnode)
	walk = func(n *ASTnode) {
		if n.Symbol == "exprseq" {
			if len(n.Children) > 0 {
				exprTemp := genExpr(&n.Children[0])
				result = append(result, exprTemp)
			}
			if len(n.Children) > 1 {
				// recurse into exprseq'
				walk(&n.Children[1])
			}
		} else if n.Symbol == "exprseq'" {
			if len(n.Children) > 1 {
				// Recurse into next exprseq
				walk(&n.Children[1])
			}
		}
	}
	walk(exprSeqNode)
	return result
}

func getVarName(varNode *ASTnode) string {
	idNode := varNode.Children[0]
	return idNode.Children[0].Lexeme
}
