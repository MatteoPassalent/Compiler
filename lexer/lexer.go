package lexer

import (
	"fmt"
	"log"
	"os"
)

type Token struct {
	Type   string
	Lexeme string
	Symbol string
	Line   int
}

func (token *Token) init(tokenType string, lexeme string, line int) {
	token.Type = tokenType
	token.Lexeme = lexeme
	token.Line = line
	if tokenType == "INTEGER" || tokenType == "DOUBLE" || tokenType == "IDENTIFIER" {
		token.Symbol = tokenType
	} else {
		token.Symbol = lexeme
	}
}

const BUFFER_SIZE = 1024
const START = 0
const PUNCTUATION = 1
const OPERATOR_1 = 2
const OPERATOR_2 = 3
const OPERATOR_3 = 4
const IDENTIFIER = 5
const INTEGER = 6
const DOUBLE_1 = 7
const DOUBLE_2 = 8
const DOUBLE_A1 = 9
const DOUBLE_3 = 10
const DOUBLE_A2 = 11
const ERROR = 12

// Initialize file pointers, buffers, and buffer variables
var (
	file              *os.File
	lexicalOutputFile *os.File
	lexicalErrorFile  *os.File
	buffer1           = make([]byte, BUFFER_SIZE+1)
	buffer2           = make([]byte, BUFFER_SIZE+1)
	currentBuffer     = buffer1
	bufferFlag        = 1
	bufferIndex       int
	bytesRead         int
	LineNumber        = 1
	ColumnNumber      int
	IsPanicMode       = false
	isInputTableInit  = false
)

// Initialize transition table from DFA
var transitionTable = [13][12]int{
	// space Punctuation    +|-         %|*|/       <           >           =      letters - e    0-9        e             .       other
	{START, PUNCTUATION, OPERATOR_3, OPERATOR_3, OPERATOR_1, OPERATOR_2, OPERATOR_2, IDENTIFIER, INTEGER, IDENTIFIER, PUNCTUATION, ERROR}, // Start
	{START, START, START, START, START, START, START, START, START, START, START, START},                                                  // Punctuation
	{START, START, START, START, START, OPERATOR_3, OPERATOR_3, START, START, START, START, START},                                        // OPERATOR_1
	{START, START, START, START, START, START, OPERATOR_3, START, START, START, START, START},                                             // OPERATOR_2
	{START, START, START, START, START, START, START, START, START, START, START, START},                                                  // OPERATOR_3
	{START, START, START, START, START, START, START, IDENTIFIER, IDENTIFIER, IDENTIFIER, START, START},                                   // IDENTIFIER
	{START, START, START, START, START, START, START, START, INTEGER, DOUBLE_2, DOUBLE_1, START},                                          // INTEGER
	{ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A1, ERROR, ERROR, ERROR},                                              // DOUBLE_1
	{ERROR, ERROR, DOUBLE_3, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                           // DOUBLE_2
	{START, START, START, START, START, START, START, START, DOUBLE_A1, DOUBLE_2, START, START},                                           // DOUBLE_A1
	{ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                              // Double_3
	{START, START, START, START, START, START, START, START, DOUBLE_A2, START, START, START},                                              // DOUBLE_A2
	{ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR},                                                  // ERROR
}

// Initialize state names for writing output
var stateNames = [13]string{
	"START", "PUNCTUATION", "OPERATOR", "OPERATOR", "OPERATOR", "IDENTIFIER", "INTEGER", "D_REJECT", "D_REJECT", "DOUBLE", "D_REJECT", "DOUBLE", "ERROR"}

// Initialize sorted keywords for binary search
var keywords = [16]string{
	"and", "def", "do", "double", "else", "fed", "fi",
	"if", "int", "not", "od", "or", "print", "return", "then", "while"}

var inputTypeTable [256]int // Initialize lookup table for input types

/*
Initializes the input type lookup table
This table is used to find the type of input character
based on its ASCII value in O(1) time
*/
func initializeInputTypeTable() {
	// Default to other (Invalid) type
	for i := 0; i < 256; i++ {
		inputTypeTable[i] = 11
	}

	// Whitespace characters
	inputTypeTable[' '] = 0
	inputTypeTable['\t'] = 0
	inputTypeTable['\n'] = 0
	inputTypeTable['\r'] = 0

	// Punctuation
	inputTypeTable['('] = 1
	inputTypeTable[')'] = 1
	inputTypeTable['['] = 1
	inputTypeTable[']'] = 1
	inputTypeTable[','] = 1
	inputTypeTable[';'] = 1

	// Operators
	inputTypeTable['+'] = 2
	inputTypeTable['-'] = 2
	inputTypeTable['*'] = 3
	inputTypeTable['/'] = 3
	inputTypeTable['%'] = 3
	inputTypeTable['<'] = 4
	inputTypeTable['>'] = 5
	inputTypeTable['='] = 6

	// Letters
	for character := 'A'; character <= 'Z'; character++ {
		inputTypeTable[character] = 7
	}
	for character := 'a'; character <= 'z'; character++ {
		inputTypeTable[character] = 7
	}

	// Digits
	for character := '0'; character <= '9'; character++ {
		inputTypeTable[character] = 8
	}

	inputTypeTable['e'] = 9 // Special case for 'e'

	inputTypeTable['.'] = 10 // Dot
}

/*
Returns the input type of a character
based on the input type lookup table
*/
func getInputType(c rune) int {
	if !isInputTableInit {
		initializeInputTypeTable()
		isInputTableInit = true
	}
	return inputTypeTable[c]
}

/*
Checks if a token is a keyword using binary search
on the sorted keywords array - O(log n) time
*/
func isKeyword(token string) bool {
	left, right := 0, len(keywords)-1

	for left <= right {
		mid := (left + right) / 2

		if token == keywords[mid] {
			return true
		} else if token < keywords[mid] {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return false
}

/*
Fills the double buffer with data from
the file. Returns the number of bytes read.
*/
func fillBuffer() int {
	if bufferFlag == 1 {
		bytesRead, _ = file.Read(buffer2[:BUFFER_SIZE])
		currentBuffer = buffer2
		bufferFlag = 2
	} else {
		bytesRead, _ = file.Read(buffer1[:BUFFER_SIZE])
		currentBuffer = buffer1
		bufferFlag = 1
	}
	bufferIndex = 0
	return bytesRead
}

/*
Returns the next character from the buffer.
If the buffer is empty, it fills the buffer
and returns the first character.
Increments the line number if the character is '\n'.
*/
func getNextChar() rune {
	if bufferIndex >= bytesRead {
		if fillBuffer() == 0 {
			return -1
		}
	}
	c := rune(currentBuffer[bufferIndex])
	bufferIndex++
	ColumnNumber++
	if c == '\n' {
		ColumnNumber = 0
		LineNumber++
	}
	return c
}

/*
Returns the next token from the buffer
based on the DFA transition table
*/
func GetNextToken() Token {
	var tokenVal string
	state := START
	for {
		c := getNextChar()
		if c == -1 { // End of file
			// If the last token ends in a non-acceptance state, it is an error token
			if state == DOUBLE_1 || state == DOUBLE_2 || state == DOUBLE_3 {
				token := Token{}
				token.init("ERROR", tokenVal, LineNumber)
				writeError(token)
				return token
				// If the last token ends in the start state, there are no more tokens
			} else if state == START {
				token := Token{}
				token.init("EOF", tokenVal, LineNumber)
				return token
			}
			// If the last token ends in an acceptance state, return the token
			token := Token{}
			token.init(stateNames[state], tokenVal, LineNumber)
			writeToken(token)
			return token
		}

		nextState := transitionTable[state][getInputType(c)]
		if nextState == ERROR { // Invalid token
			tokenVal += string(c)
			token := Token{}
			token.init("ERROR", tokenVal, LineNumber)
			writeError(token)
			return token
		} else if nextState == START { // Transistioned back to start state
			if state == START { // (Start -> Start) Skip whitespace
				continue
			}
			// Token ended: Transistioned from an acceptance state to start state for the next token
			bufferIndex-- // Move back one character for next token
			if c == '\n' {
				LineNumber-- // Decrement line number if next char is newline character
			} else {
				ColumnNumber-- // Decrement column number if next char is not newline character
			}
			if state == IDENTIFIER && isKeyword(tokenVal) {
				token := Token{}
				token.init("KEYWORD", tokenVal, LineNumber)
				writeToken(token)
				return token
			}
			token := Token{}
			token.init(stateNames[state], tokenVal, LineNumber)
			writeToken(token)
			return token
		} else { // Continued transition - Add character to token
			tokenVal += string(c)
		}
		state = nextState
	}
}

func ErrorRecovery() {
	IsPanicMode = true
	for {
		token := GetNextToken()
		if token.Type == "EOF" {
			break
		}
	}
}

func writeToken(token Token) {
	fmt.Fprintf(lexicalOutputFile, "Type: %-15s Token: %s\n", token.Type, token.Lexeme)
}

func writeError(token Token) {
	fmt.Fprintf(lexicalErrorFile, "Error (Line %d, Column %d): Invalid token %s\n", LineNumber, ColumnNumber, token.Lexeme)
}

func InitLexerFiles(sourceFile *os.File) {
	var err error
	file = sourceFile
	lexicalOutputFile, err = os.Create("lexer/lexical_output.txt")
	if err != nil {
		log.Fatal("Error creating output file:", err)
	}
	lexicalErrorFile, err = os.Create("lexer/lexical_errors.txt")
	if err != nil {
		log.Fatal("Error creating error file:", err)
	}
}

func CloseLexerFiles() {
	defer file.Close()
	defer lexicalOutputFile.Close()
	defer lexicalErrorFile.Close()
}
