package codegen

import (
	"Compiler/parser"
	"fmt"
)

// We can define a structure to hold a single 3-address (or similar) instruction.
type TACInstruction struct {
	// For simplicity, store each instruction in a single string or in fields
	Op   string // the operation, e.g., "ADD", "SUB", "MUL", "CALL", etc.
	Arg1 string
	Arg2 string
	Arg3 string
}

// We'll store the generated instructions in a list.
var instructions []TACInstruction

// For naming temporaries & labels
var tempCount int
var labelCount int

// For tracking the current function name
var currentFunctionName string

// MakeTemp generates a new temporary variable name.
func MakeTemp() string {
	tempCount++
	return fmt.Sprintf("t%d", tempCount)
}

// MakeLabel generates a new label name.
func MakeLabel(prefix string) string {
	labelCount++
	return fmt.Sprintf("%s%d", prefix, labelCount)
}

// Emit appends a new instruction to our 3TAC instruction list.
func Emit(op, arg1, arg2, dest string) {
	instructions = append(instructions, TACInstruction{op, arg1, arg2, dest})
}

// GenerateCode is the main entry point for the code generator.
// It returns a list of generated 3TAC instructions.
func GenerateCode(ast *parser.ASTnode) []TACInstruction {
	instructions = []TACInstruction{}
	tempCount = 0
	labelCount = 0

	// Begin code gen from the root of the AST
	genProgram(ast)
	for _, instruction := range instructions {
		fmt.Printf("%s %s %s %s\n", instruction.Op, instruction.Arg1, instruction.Arg2, instruction.Arg3)
	}
	return instructions
}

// genProgram handles the top-level "program" node
// You can adapt based on how your root AST is structured.
func genProgram(node *parser.ASTnode) {
	// Typically, you might have children representing fdecls, declarations, statement_seq, etc.
	// We just walk them in order:
	Emit("B", "main", "", "")

	for _, child := range node.Children {
		switch child.Symbol {
		case "fdecls", "fdecls'":
			genFdecls(&child)
		case "declarations", "declarations'":
			// Possibly global declarations, etc.
			// (Often you might not generate code for these, just store them in the symbol table.)
		case "statement_seq":
			// The "main" portion or top-level statements
			Emit("main:", "", "", "")
			genStatementSeq(&child)
		}
	}
}

// genFdecls walks a list of function declarations
func genFdecls(node *parser.ASTnode) {
	for _, child := range node.Children {
		if child.Symbol == "fdec" {
			genFdec(&child)
		} else if child.Symbol == "fdecls'" {
			// Or if child is fdecls' again, call genFdecls recursively
			genFdecls(&child)
		}
	}
}

// genFdec handles a single function definition: def type id (params) declarations statement_seq fed
func genFdec(node *parser.ASTnode) {
	// child layout may differ in your AST
	// e.g. node.Children[0] = "def", node.Children[1] = type, node.Children[2] = id, etc.
	// Adjust indexing as needed.

	funcNameNode := node.Children[2] //  "id"
	funcName := funcNameNode.Children[0].Lexeme

	currentFunctionName = funcName
	// Make a label for the function
	Emit(fmt.Sprintf("%s:", funcName), "", "", "") // e.g. "gcd:"
	Emit("Begin", "", "", "")                      // Could push {LR}, push {FP}, etc.
	Emit("Push", "{LR}", "", "")                   // Save the old link register
	Emit("Push", "{FP}", "", "")                   // Save the old frame pointer

	fpCounter := 4
	// Possibly handle the declarations child (locals) or do nothing if you store them in a table

	paramsNode := node.Children[4]
	var processParams func(*parser.ASTnode)
	processParams = func(pNode *parser.ASTnode) {
		if pNode == nil || len(pNode.Children) < 3 {
			return
		}

		// Extract identifier
		idNode := pNode.Children[1].Children[0] // id
		paramName := idNode.Children[0].Lexeme  // IDENTIFIER

		fpCounter += 4
		Emit("ADD", paramName, "{FP}", fmt.Sprintf("%d", fpCounter)) // e.g. "ADD {FP}, 4, paramName"

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
	// Now handle the statements inside this function
	for _, child := range node.Children {
		switch child.Symbol {
		case "statement_seq":
			genStatementSeq(&child)
		}
	}

	Emit(currentFunctionName+"Exit:", "", "", "") // Label for the exit point
	Emit("Pop", "{FP}", "", "")                   // Restore the old frame pointer
	Emit("Pop", "{PC}", "", "")                   // Restore the old link register
	currentFunctionName = ""
}

func genStatementSeq(node *parser.ASTnode) {
	// Safety checks
	if node == nil || node.Symbol != "statement_seq" {
		return
	}
	// Grammar: statement_seq -> statement statement_seq'
	// child[0] = statement
	// child[1] = statement_seq'
	if len(node.Children) == 0 {
		return
	}

	// 1) Generate code for the first statement
	statementNode := &node.Children[0]
	genStatement(statementNode)

	// 2) If there's a second child, it should be statement_seq'
	if len(node.Children) > 1 {
		statementSeqPrime := &node.Children[1]
		genStatementSeqPrime(statementSeqPrime)
	}
}

func genStatementSeqPrime(node *parser.ASTnode) {
	// Grammar: statement_seq' -> ; statement_seq | ε
	// If it's empty, node.Children might be 0.
	if node == nil || node.Symbol != "statement_seq'" {
		return
	}
	if len(node.Children) == 0 {
		// It's epsilon (empty)
		return
	}

	// If not empty, then:
	// child[0] = ";" punctuation
	// child[1] = statement_seq
	semicolonNode := &node.Children[0]
	if semicolonNode.Lexeme == ";" {
		statementSeqNode := &node.Children[1]
		genStatementSeq(statementSeqNode)
	}
}

// genStatement handles a single statement node
func genStatement(node *parser.ASTnode) {
	// Your AST might store statements in different ways. The gist:
	// 1) if it's an assignment: var = expr
	// 2) if it's if bexpr then statement_seq statement'
	// 3) if it's while ...
	// 4) if it's print ...
	// 5) if it's return ...
	// etc.

	// This is a simplified demonstration. Adjust to your actual node structure.

	// Check the first child or the structure to see which statement it is.
	if len(node.Children) == 0 {
		// Empty statement
		return
	}

	// For example, detect assignment: statement -> var = expr
	if len(node.Children) >= 3 && node.Children[1].Symbol == "=" {
		// var = expr
		varNode := &node.Children[0]
		exprNode := &node.Children[2]

		// Generate code for the expression
		exprResult := genExpr(exprNode)

		// The varNode might be something like var -> id or var -> id [ expr ]
		// We'll keep it simple: if it's just an identifier
		varName := getVarName(varNode)
		Emit("MOV", exprResult, varName, "")
		return
	}

	// Possibly an "if" statement
	if node.Children[0].Symbol == "if" {
		// if bexpr then statement_seq statement'
		bexprNode := &node.Children[1]
		trueSeqNode := &node.Children[3] // statement_seq
		// statement' might contain else or fi

		elseOrFiNode := &node.Children[4]

		// labelElse := MakeLabel("Lelse")
		// labelEnd := MakeLabel("LendIf")
		labelTrue := MakeLabel("L")
		labelFalse := MakeLabel("L")

		// Generate boolean expression code
		condResult := genBexpr(bexprNode)
		// We can generate a jump-if-false:
		// In typical 3AC, you might do: if condResult == 0 goto labelElse
		Emit("B"+condResult, labelTrue, "", "")
		Emit("B", labelFalse, "", "") // Jump to false block
		// Emit("IF_FALSE_GOTO", condResult, "", labelElse)

		// True block
		Emit(labelTrue+":", "", "", "")
		genStatementSeq(trueSeqNode)

		// Jump to end
		// Emit("GOTO", labelEnd, "", "")

		// Else part
		// Emit("", "", "", labelElse+":")
		// Could be "else statement_seq fi" or just "fi"
		Emit(labelFalse+":", "", "", "")
		genElsePart(elseOrFiNode)
		// Emit("", "", "", labelEnd+":")
		return
	}

	// Possibly a "while" statement
	if node.Children[0].Symbol == "while" {
		condNode := &node.Children[1]
		bodyNode := &node.Children[3]

		loopLabel := MakeLabel("L")
		bodyLabel := MakeLabel("L")
		exitLabel := MakeLabel("L")

		Emit(loopLabel+":", "", "", "") // Loop start
		condResult := genBexpr(condNode)

		Emit("B"+condResult, bodyLabel, "", "") // Jump to body if true
		Emit("B", exitLabel, "", "")            // Jump to exit if false
		Emit(bodyLabel+":", "", "", "")         // Body start
		genStatementSeq(bodyNode)
		Emit("B", loopLabel, "", "")    // Jump back to loop start
		Emit(exitLabel+":", "", "", "") // Exit label

		return
	}

	// Possibly a "print expr"
	if node.Children[0].Symbol == "print" {
		exprNode := &node.Children[1]
		exprResult := genExpr(exprNode)
		Emit("PRINT", exprResult, "", "")
		return
	}

	// Possibly a "return expr"
	if node.Children[0].Symbol == "return" {
		exprNode := &node.Children[1]
		exprResult := genExpr(exprNode)
		// We just store it in a hidden location or a standard place:
		Emit("MOV", "{fp - 4}", exprResult, "")
		Emit("B", currentFunctionName+"Exit", "", "") // Return to caller
		return
	}

	// ... handle other statement forms ...
}

// genElsePart tries to figure out if there's an "else" in statement'
func genElsePart(node *parser.ASTnode) {
	// statement' -> fi OR else statement_seq fi
	// If no children or the first child is "fi", there's no else
	if len(node.Children) > 0 && node.Children[0].Symbol == "else" {
		seqNode := &node.Children[1] // statement_seq
		genStatementSeq(seqNode)
	}
}

func genBexpr(node *parser.ASTnode) string {
	// bexpr → bterm bexpr'
	if node == nil || node.Symbol != "bexpr" || len(node.Children) == 0 {
		return ""
	}

	left := genBterm(&node.Children[0]) // first bterm

	if len(node.Children) == 1 { // ε
		return left
	}
	return genBexprPrime(&node.Children[1], left)
}

// bexpr' → or bterm bexpr' | ε
func genBexprPrime(node *parser.ASTnode, leftTmp string) string {
	if node == nil || len(node.Children) == 0 {
		return leftTmp // ε
	}

	// child[0]  = 'or'
	// child[1]  = bterm
	// child[2]  = bexpr'
	// opNode := node.Children[0] // keyword "or"
	rightTmp := genBterm(&node.Children[1])

	out := MakeTemp()
	Emit("OR", leftTmp, rightTmp, out) // out = leftTmp OR rightTmp

	if len(node.Children) > 2 { // still more ors
		return genBexprPrime(&node.Children[2], out)
	}
	return out
}

// bterm → bfactor bterm'
func genBterm(node *parser.ASTnode) string {
	if node == nil || node.Symbol != "bterm" || len(node.Children) == 0 {
		return ""
	}
	left := genBfactor(&node.Children[0])

	if len(node.Children) == 1 {
		return left
	}
	return genBtermPrime(&node.Children[1], left)
}

// bterm' → and bfactor bterm' | ε
func genBtermPrime(node *parser.ASTnode, leftTmp string) string {
	if node == nil || len(node.Children) == 0 {
		return leftTmp
	}

	// child[0] = 'and'
	// child[1] = bfactor
	rightTmp := genBfactor(&node.Children[1])

	out := MakeTemp()
	Emit("AND", leftTmp, rightTmp, out) // out = leftTmp AND rightTmp

	if len(node.Children) > 2 {
		return genBtermPrime(&node.Children[2], out)
	}
	return out
}

// bfactor → not bfactor
//
//	| ( expr comp expr )
func genBfactor(node *parser.ASTnode) string {
	if node == nil || node.Symbol != "bfactor" {
		return ""
	}

	// Case 1: 'not' bfactor
	if node.Children[0].Symbol == "not" {
		sub := genBfactor(&node.Children[1])
		out := MakeTemp()
		Emit("NOT", sub, "", out) // out = NOT sub
		return out
	}

	// Case 2: ( expr comp expr )
	// Layout (see sample tree):
	// child[0] = '('
	// child[1] = expr
	// child[2] = comp
	// child[3] = expr
	// child[4] = ')'
	if len(node.Children) >= 5 && node.Children[0].Lexeme == "(" {
		left := genExpr(&node.Children[1])
		opTmp := node.Children[2] // comp node
		right := genExpr(&node.Children[3])

		// Extract actual comparison operator symbol:
		compTok := opTmp.Children[0].Lexeme // <  >  ==  <= ...
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

		Emit("CMP", left, right, "") // out = (left compTok right) ?1:0
		return compTok
	}

	// If grammar later adds literals/identifiers in bfactor, handle here.
	return ""
}

func genExpr(node *parser.ASTnode) string {
	// node should be: expr -> term expr'
	// child[0] is the 'term'
	// child[1] is the 'expr' (or 'expr' prime) part
	if len(node.Children) == 2 && node.Symbol == "expr" {
		// Generate code for the term
		leftTemp := genTerm(&node.Children[0])
		// Now handle expr'
		return genExprPrime(&node.Children[1], leftTemp)
	}

	// Fallback if your AST is slightly different or if it's just a single child
	if len(node.Children) == 1 {
		return genExpr(&node.Children[0])
	}

	return ""
}

func genExprPrime(node *parser.ASTnode, leftTemp string) string {
	// node should be expr' -> + term expr'
	//                   or -> - term expr'
	//                   or -> ε  (empty)

	// If expr' is empty (epsilon), just return leftTemp
	// Typically, an empty node might have 0 children or a special symbol.
	if len(node.Children) == 0 {
		return leftTemp
	}

	// If we do have children, they might be:
	// child[0] = '+' or '-'
	// child[1] = 'term'
	// child[2] = 'expr' (prime again)
	op := node.Children[0].Symbol
	termNode := &node.Children[1]
	rightTemp := genTerm(termNode)

	// Now create a new temp to combine leftTemp op rightTemp
	newTemp := MakeTemp()
	switch op {
	case "+":
		Emit("ADD", newTemp, leftTemp, rightTemp)
	case "-":
		Emit("SUB", newTemp, leftTemp, rightTemp)
		// If your grammar includes more operators at expr' level, handle them
	}

	// Recurse on the remaining expr'
	if len(node.Children) > 2 {
		return genExprPrime(&node.Children[2], newTemp)
	}

	return newTemp
}

func genTerm(node *parser.ASTnode) string {
	// term -> factor term'
	// child[0] = factor
	// child[1] = term'
	if len(node.Children) == 2 && node.Symbol == "term" {
		leftTemp := genFactor(&node.Children[0])
		return genTermPrime(&node.Children[1], leftTemp)
	}

	// fallback
	if len(node.Children) == 1 {
		return genTerm(&node.Children[0])
	}
	return ""
}

func genTermPrime(node *parser.ASTnode, leftTemp string) string {
	// term' -> * factor term'
	//        | / factor term'
	//        | % factor term'
	//        | ε

	// if epsilon
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

func genFactor(node *parser.ASTnode) string {

	if isCall, idN, exprSeqN := detectCall(node); isCall {
		return genFunctionCall(*idN, exprSeqN)
	}
	// If factor -> INTEGER
	if len(node.Children) == 1 && node.Children[0].Symbol == "INTEGER" {
		val := node.Children[0].Lexeme
		tmp := MakeTemp()
		Emit("MOV", val, tmp, "")
		return tmp
	}

	// If factor -> DOUBLE
	if len(node.Children) == 1 && node.Children[0].Symbol == "DOUBLE" {
		val := node.Children[0].Lexeme
		tmp := MakeTemp()
		Emit("MOV", val, tmp, "")
		return tmp
	}

	// If factor -> IDENTIFIER (variable)
	if node.Children[0].Symbol == "id" {
		return node.Children[0].Children[0].Lexeme
	}
	// If factor -> ( expr )
	// child[0] = "("
	// child[1] = expr
	// child[2] = ")"
	if len(node.Children) == 3 && node.Children[0].Lexeme == "(" {
		return genExpr(&node.Children[1])
	}

	// If factor -> function call or "id factor'", handle that
	// ...

	// If your grammar nest is deeper, keep drilling down
	// or fallback
	if len(node.Children) == 1 {
		return genFactor(&node.Children[0])
	}

	return ""
}

// returns (isCall, idNode, exprSeqNode)
func detectCall(factor *parser.ASTnode) (bool, *parser.ASTnode, *parser.ASTnode) {
	if factor == nil || factor.Symbol != "factor" || len(factor.Children) < 2 {
		return false, nil, nil
	}
	idNode := &factor.Children[0]       // id
	factorPrime := &factor.Children[1]  // factor'
	if len(factorPrime.Children) > 0 && // must have "("
		factorPrime.Children[0].Lexeme == "(" {
		// exprseq might be empty, but grammar still gives the node
		exprSeq := &factorPrime.Children[1] // exprseq
		return true, idNode, exprSeq
	}
	return false, nil, nil
}

// genFunctionCall handles calls like gcd( exprseq )
func genFunctionCall(idNode parser.ASTnode, exprSeqNode *parser.ASTnode) string {
	funcName := idNode.Children[0].Lexeme // the actual IDENTIFIER from id->identifier
	// Evaluate each arg in exprseq
	argTemps := collectArgs(exprSeqNode)

	// In many 3AC forms, we might do "PARAM" instructions or "push" instructions:
	for _, arg := range argTemps {
		Emit("PUSH", arg, "", "")
	}

	// Then call the function
	returnTemp := MakeTemp()
	Emit("BL", funcName, returnTemp, "") // BL = branch and link (call)

	// Pop the arguments (depending on your calling convention)
	for i := len(argTemps) - 1; i >= 0; i-- {
		arg := argTemps[i]
		Emit("POP", arg, "", "")
	}

	return returnTemp
}

// collectArgs walks exprseq to produce each argument
func collectArgs(exprSeqNode *parser.ASTnode) []string {
	// exprseq -> expr exprseq' or empty
	// exprseq' -> , exprseq or empty
	var result []string

	var walk func(*parser.ASTnode)
	walk = func(n *parser.ASTnode) {
		if n.Symbol == "exprseq" {
			if len(n.Children) > 0 {
				// first child is an expr
				exprTemp := genExpr(&n.Children[0])
				result = append(result, exprTemp)
			}
			if len(n.Children) > 1 {
				// second child is exprseq'
				walk(&n.Children[1])
			}
		} else if n.Symbol == "exprseq'" {
			if len(n.Children) > 1 {
				// child[1] is next exprseq
				walk(&n.Children[1])
			}
		}
	}

	walk(exprSeqNode)
	return result
}

// // Helper to get the name of a var node, adjusting if it's array-like
// func getVarName(varNode *parser.ASTnode) string {
// 	// var -> id var'
// 	// If var' is empty, it's just the simple id
// 	idNode := varNode.Children[0]
// 	baseName := idNode.Children[0].Lexeme // IDENTIFIER text

//		// If var' is [ expr ], we have an array element
//		if len(varNode.Children) > 1 {
//			varPrime := varNode.Children[1]
//			if len(varPrime.Children) > 0 && varPrime.Children[0].Symbol == "[" {
//				// array indexing
//				idx := genExpr(&varPrime.Children[1])
//				// For simplicity in 3TAC, we might do a separate temp = baseName[idx]
//				t := MakeTemp()
//				Emit("LOAD_ARR", baseName, idx, t)
//				// We could return t as the “location”, or you might do something else
//				return t
//			}
//		}
//		return baseName
//	}
//
// TODO: Those bits somehow
func getVarName(varNode *parser.ASTnode) string {
	// If var -> id var'
	// child[0] is `id`
	idNode := varNode.Children[0]
	// idNode.Children[0] is `IDENTIFIER`
	return idNode.Children[0].Lexeme
}
