// Once a final acceptance state is reached where the next token does not match anything, then
// Transisiton to TERMINATE state, program checks if TERMINATE first thing, if so emits last state and
// token - most recent char, resets state to start and token to the most recent char
// Can prob just get rid of terminate and go straight to start

// can do binary seach on comparing keywords to identifier

// For keywords I think we can just check direcly. So when trasisiton table terminates with identifier just check if it == a keyword and if so emit keyword token
// just have to write my own !strcmp function to compare the strings but otherwise allowed I think

// ERROR STATE needs to do one thing if the current char is valid on its own and another if its not

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
#define TERMINATE 13

char buffer1[BUFFER_SIZE + 1], buffer2[BUFFER_SIZE + 1];
char *currentBuffer = buffer1;
int bufferIndex = 0, bytesRead = 0;
int lineNumber = 1;
FILE *file, *outputFile, *errorFile;

int transition_table[14][12] = {
    // space  Punctuation    +|-         %|*|/      <           >             =      letters - e      0-9      e         .       other
    {START, PUNCTUATION, OPERATOR_3, OPERATOR_3, OPERATOR_1, OPERATOR_2, OPERATOR_2, IDENTIFIER, INTEGER, IDENTIFIER, ERROR, ERROR},         // Start
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE},    // Punctuation
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, OPERATOR_3, OPERATOR_3, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE},  // OPERATOR_1
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, OPERATOR_3, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE},   // OPERATOR_2
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE},    // OPERATOR_3
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, IDENTIFIER, IDENTIFIER, IDENTIFIER, TERMINATE, TERMINATE}, // IDENTIFIER
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, INTEGER, DOUBLE_2, DOUBLE_1, TERMINATE},        // INTEGER
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A1, ERROR, ERROR, ERROR},                                                // DOUBLE_1
    {ERROR, ERROR, DOUBLE_3, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                             // DOUBLE_2
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, DOUBLE_A1, DOUBLE_2, TERMINATE, TERMINATE},     // DOUBLE_A1
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, DOUBLE_A2, ERROR, ERROR, ERROR},                                                // Double_3
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, DOUBLE_A2, TERMINATE, TERMINATE, TERMINATE},               // DOUBLE_A2
    {ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR, ERROR},                                                    // ERROR
    {TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE, TERMINATE},    // TERMINATE
};

char *state_names[14] = {
    "REJECT", "PUNCTUATION", "OPERATOR", "OPERATOR", "OPERATOR", "IDENTIFIER", "INTEGER", "REJECT", "REJECT", "DOUBLE", "REJECT", "DOUBLE", "REJECT", "REJECT"};

const char *keywords[] = {
    "and", "def", "do", "double", "else", "fed", "fi",
    "if", "int", "not", "od", "or", "print", "return", "then", "while"};

int isLetter(char c)
{
    return ((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'));
}

int isDigit(char c)
{
    return (c >= '0' && c <= '9');
}

int getInputType(char c) // TODO: check if more efficient way
{
    if (c == ' ' || c == '\t' || c == '\n' || c == '\r')
        return 0;
    if (c == '(' || c == ')' || c == '[' || c == ']' || c == ',' || c == ';')
        return 1;
    if (c == '+' || c == '-')
        return 2;
    if (c == '*' || c == '/' || c == '%')
        return 3;
    if (c == '<')
        return 4;
    if (c == '>')
        return 5;
    if (c == '=')
        return 6;
    if (isLetter(c) && c != 'e')
        return 7;
    if (isDigit(c))
        return 8;
    if (c == 'e')
        return 9;
    if (c == '.')
        return 10;
    return 11;
}

int compareStrings(const char *s1, const char *s2)
{
    while (*s1 && *s2 && (*s1 == *s2))
    {
        s1++;
        s2++;
    }
    return (unsigned char)*s1 - (unsigned char)*s2;
}

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

char getNextChar()
{
    if (bufferIndex >= bytesRead)
    {
        if (fillBuffer() == 0)
            return '\0';
    }
    char c = currentBuffer[bufferIndex++];
    if (c == '\n')
        lineNumber++;
    return c;
}

int getNextToken(char *token, int *state)
{
    int tokenIndex = 0;
    while (1)
    {
        char c = getNextChar();
        if (c == '\0')
        {
            if (*state == DOUBLE_1 || *state == DOUBLE_2 || *state == DOUBLE_3)
            {
                token[tokenIndex] = '\0';
                *state = ERROR;
                return 1;
            }
            else if (*state != START)
            {
                token[tokenIndex] = '\0';
                return 0;
            }
            else
                return 1;
        }

        int inputType = getInputType(c);
        int nextState = transition_table[*state][inputType];

        if (nextState == ERROR)
        {
            token[tokenIndex++] = c;
            token[tokenIndex] = '\0';
            *state = ERROR;
            return 1;
        }
        else if (nextState == TERMINATE) // TODO: Try getting rid of terminate
        {
            token[tokenIndex] = '\0';
            bufferIndex--;
            if (c == '\n')
                lineNumber--;
            return 0;
        }
        else if (nextState == START)
        {
            tokenIndex = 0;
        }
        else
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

        if (isValidToken == 0)
        {
            if (state == IDENTIFIER && isKeyword(token))
            {
                fprintf(outputFile, "Type: %-15s Token: %s\n", "KEYWORD", token);
            }
            else
            {
                fprintf(outputFile, "Type: %-15s Token: %s\n", state_names[state], token);
            }
        }
        else if (isValidToken == 1)
        {
            if (state == ERROR)
                fprintf(errorFile, "Error (Line %d): Invalid token %s\n", lineNumber, token);
            else
                break;
        }
    }
}

// Main Function
int main()
{
    file = fopen("./Testing/Test7.cp", "r");
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

    fillBuffer();
    lexicalAnalysis();
    fclose(file);

    return 0;
}
