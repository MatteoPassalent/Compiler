#include <stdio.h>

#define BUFFER_SIZE 1024
#define START 0
#define PUNCTUATION 1
#define OPERATOR_1 2
#define OPERATOR_2 3
#define OPERATOR_3 4
#define IDENTIFIER 5
#define INTEGER 6
#define DOUBLE_1 7
#define DOUBLE_2 8
#define DOUBLE_A1 9
#define DOUBLE_3 10
#define DOUBLE_A2 11
#define ERROR 12

// Initialize file pointers, buffers, and buffer variables
FILE *file, *outputFile, *errorFile;
char buffer1[BUFFER_SIZE + 1], buffer2[BUFFER_SIZE + 1];
char *currentBuffer = buffer1;
int bufferIndex = 0, bytesRead = 0;

int lineNumber = 1;
int columnNumber = 0;

int inputTypeTable[256]; // Initialize lookup table for input types

// Initialize transition table from DFA
int transition_table[13][12] = {
    // space Punctuation    +|-         %|*|/       <           >           =      letters - e    0-9        e         .     other
    {START, PUNCTUATION, OPERATOR_3, OPERATOR_3, OPERATOR_1, OPERATOR_2, OPERATOR_2, IDENTIFIER, INTEGER, IDENTIFIER, ERROR, ERROR}, // Start
    {START, START, START, START, START, START, START, START, START, START, START, START},                                            // Punctuation
    {START, START, START, START, START, OPERATOR_3, OPERATOR_3, START, START, START, START, START},                                  // OPERATOR_1
    {START, START, START, START, START, START, OPERATOR_3, START, START, START, START, START},                                       // OPERATOR_2
    {START, START, START, START, START, START, START, START, START, START, START, START},                                            // OPERATOR_3
    {START, START, START, START, START, START, START, IDENTIFIER, IDENTIFIER, IDENTIFIER, START, START},                             // IDENTIFIER
    {START, START, START, START, START, START, START, START, INTEGER, DOUBLE_2, DOUBLE_1, START},                                    // INTEGER
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A1, ERROR, ERROR, ERROR},                                        // DOUBLE_1
    {ERROR, ERROR, DOUBLE_3, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                     // DOUBLE_2
    {START, START, START, START, START, START, START, START, DOUBLE_A1, DOUBLE_2, START, START},                                     // DOUBLE_A1
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                        // Double_3
    {START, START, START, START, START, START, START, START, DOUBLE_A2, START, START, START},                                        // DOUBLE_A2
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR},                                            // ERROR
};

// Initialize state names for writing output
char *state_names[13] = {
    "START", "PUNCTUATION", "OPERATOR", "OPERATOR", "OPERATOR", "IDENTIFIER", "INTEGER", "D_REJECT", "D_REJECT", "DOUBLE", "D_REJECT", "DOUBLE", "ERROR"};

// Initialize sorted keywords for binary search
const char *keywords[] = {
    "and", "def", "do", "double", "else", "fed", "fi",
    "if", "int", "not", "od", "or", "print", "return", "then", "while"};

/*
    Initializes the input type lookup table
    This table is used to find the type of input character
    based on its ASCII value in O(1) time
*/
void initializeInputTypeTable()
{
    // Default to other (Invalid) type
    for (int i = 0; i < 256; i++)
        inputTypeTable[i] = 11;

    // Whitespace characters
    inputTypeTable[' '] = 0;
    inputTypeTable['\t'] = 0;
    inputTypeTable['\n'] = 0;
    inputTypeTable['\r'] = 0;

    // Punctuation
    inputTypeTable['('] = 1;
    inputTypeTable[')'] = 1;
    inputTypeTable['['] = 1;
    inputTypeTable[']'] = 1;
    inputTypeTable[','] = 1;
    inputTypeTable[';'] = 1;

    // Operators
    inputTypeTable['+'] = 2;
    inputTypeTable['-'] = 2;
    inputTypeTable['*'] = 3;
    inputTypeTable['/'] = 3;
    inputTypeTable['%'] = 3;
    inputTypeTable['<'] = 4;
    inputTypeTable['>'] = 5;
    inputTypeTable['='] = 6;

    // Letters
    for (char c = 'A'; c <= 'Z'; c++)
        inputTypeTable[(unsigned char)c] = 7;
    for (char c = 'a'; c <= 'z'; c++)
        inputTypeTable[(unsigned char)c] = 7;

    // Digits
    for (char c = '0'; c <= '9'; c++)
        inputTypeTable[(unsigned char)c] = 8;

    inputTypeTable['e'] = 9; // Special case for 'e'

    inputTypeTable['.'] = 10; // Dot
}

/*
    Returns the input type of a character
    based on the input type lookup table
*/
int getInputType(char c)
{
    return inputTypeTable[(unsigned char)c];
}

/*
    Compares two strings lexicographically
    Returns 0 if the strings are equal
    Returns a positive value if s1 > s2
    Returns a negative value if s1 < s2
*/
int compareStrings(const char *s1, const char *s2)
{
    while (*s1 && *s2 && (*s1 == *s2))
    {
        s1++;
        s2++;
    }
    return (unsigned char)*s1 - (unsigned char)*s2;
}

/*
    Checks if a token is a keyword using binary search
    on the sorted keywords array - O(log n) time
*/
int isKeyword(const char *token)
{
    int left = 0;
    int right = (sizeof(keywords) / sizeof(keywords[0])) - 1;

    while (left <= right)
    {
        int mid = (left + right) / 2;
        int cmp = compareStrings(token, keywords[mid]);

        if (cmp == 0)
        {
            return 1;
        }
        else if (cmp < 0)
        {
            right = mid - 1;
        }
        else
        {
            left = mid + 1;
        }
    }
    return 0;
}

/*
    Fills the double buffer with data from
    the file. Returns the number of bytes read
*/
int fillBuffer()
{
    if (currentBuffer == buffer1)
    {
        bytesRead = fread(buffer2, 1, BUFFER_SIZE, file);
        buffer2[bytesRead] = '\0';
        currentBuffer = buffer2;
    }
    else
    {
        bytesRead = fread(buffer1, 1, BUFFER_SIZE, file);
        buffer1[bytesRead] = '\0';
        currentBuffer = buffer1;
    }
    bufferIndex = 0;
    return bytesRead;
}

/*
    Returns the next character from the buffer
    If the buffer is empty, it fills the buffer
    and returns the first character
    Iterates the line number if the character is '\n'
*/
char getNextChar()
{
    if (bufferIndex >= bytesRead)
    {
        if (fillBuffer() == 0)
            return '\0';
    }
    char c = currentBuffer[bufferIndex++];
    columnNumber++;
    if (c == '\n')
    {
        columnNumber = 0;
        lineNumber++;
    }
    return c;
}

/*
    Returns the next token from the buffer
    based on the DFA transition table
*/
int getNextToken(char *token, int *state)
{
    int tokenIndex = 0;
    while (1)
    {
        char c = getNextChar();
        if (c == '\0') // End of file
        {
            // If the last token ends in a non-acceptance state, it is an error token
            if (*state == DOUBLE_1 || *state == DOUBLE_2 || *state == DOUBLE_3)
            {
                token[tokenIndex] = '\0';
                *state = ERROR;
                return 1;
            }
            // If the last token ends in the start state, there are no more tokens
            else if (*state == START)
            {
                return 1;
            }
            // If the last token ends in an acceptance state, it is a valid token
            token[tokenIndex] = '\0';
            return 0;
        }

        int inputType = getInputType(c);
        int nextState = transition_table[*state][inputType];

        if (nextState == ERROR) // Invalid token
        {
            token[tokenIndex++] = c;
            token[tokenIndex] = '\0';
            *state = ERROR;
            return 1;
        }
        else if (nextState == START) // Transistioned back to start state
        {
            if (*state == START) // (Start -> Start) Skip whitespace
            {
                tokenIndex = 0;
                continue;
            }

            // Token ended: Transistioned from an acceptance state to start state for the next token:
            token[tokenIndex] = '\0';
            bufferIndex--; // Move back one character for next token
            if (c == '\n')
                lineNumber--; // Decrement line number if next char is newline character
            else
                columnNumber--; // Decrement column number if next char is not newline character
            return 0;
        }
        else // Continued transition - Add character to token
        {
            token[tokenIndex++] = c;
        }
        *state = nextState;
    }
}

// Lexical Analyzer
void lexicalAnalysis()
{
    char token[256];
    while (1)
    {
        int state = START;
        int isValidToken = getNextToken(token, &state);

        if (isValidToken == 0) // Valid token
        {
            if (state == IDENTIFIER && isKeyword(token)) // Check if token is a keyword
            {
                fprintf(outputFile, "Type: %-15s Token: %s\n", "KEYWORD", token);
            }
            else
            {
                fprintf(outputFile, "Type: %-15s Token: %s\n", state_names[state], token);
            }
        }
        else if (isValidToken == 1) // Error token or end of file
        {
            if (state == ERROR) // If in error state -> invalid token
                fprintf(errorFile, "Error (Line %d, Column %d): Invalid token %s\n", lineNumber, columnNumber, token);
            else // If no valid token returned and not in error state -> end of file
                break;
        }
    }
}

int main(int argc, char *argv[])
{
    if (argc != 2)
    {
        fprintf(stderr, "Usage: %s <input_file>\n", argv[0]);
        return 1;
    }

    file = fopen(argv[1], "r");
    if (file == NULL)
    {
        perror("Error opening file");
        return 1;
    }

    outputFile = fopen("output.txt", "w");
    errorFile = fopen("error.txt", "w");

    if (!outputFile || !errorFile)
    {
        perror("Error opening output/error files");
        return 1;
    }

    initializeInputTypeTable();
    fillBuffer();
    lexicalAnalysis();

    fclose(file);
    fclose(outputFile);
    fclose(errorFile);

    return 0;
}
