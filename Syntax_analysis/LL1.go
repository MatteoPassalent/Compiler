package main

var ll1 = [33][39]string{
	// . ; def ( ) fed , int double = if then fi else while do od print return + - * / % or and not < > == [ ] letters digits integers doubles
	{"", "program fdecls declarations statement_seq .", "program fdecls declarations statement_seq .", "program fdecls declarations statement_seq .", "", "", "", "", "program fdecls declarations statement_seq .", "program fdecls declarations statement_seq .", "", "program fdecls declarations statement_seq .", "", "", "", "program fdecls declarations statement_seq .", "", "", "program fdecls declarations statement_seq .", "program fdecls declarations statement_seq .", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "program fdecls declarations statement_seq ."}, // Pgrm
	{"", "fdecls e", "fdecls e", "fdecls fdec ; fdecls'", "", "", "", "", "fdecls e", "fdecls e", "", "fdecls e", "", "", "", "fdecls e", "", "", "fdecls e", "fdecls e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "fdecls e"},                                                                                                                                                                                                                                                                                                                                                  // fdecls
	{"", "fdecls' e", "fdecls' e", "fdecls' fdec ; fdecls'", "", "", "", "", "fdecls' e", "fdecls' e", "", "fdecls' e", "", "", "", "fdecls' e", "", "", "fdecls' e", "fdecls' e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "fdecls' e"},                                                                                                                                                                                                                                                                                                                                        // fdecls'
	{"", "", "", "fdec def type id ( params ) declarations statement_seq fed", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                     // fdec
	{"", "", "", "", "", "params e", "", "", "params type var params'", "params type var params'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                         // params
	{"", "", "", "", "", "params' e", "", "params' , params", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                                      // params'
	{"", "declarations e", "declarations e", "", "", "", "declarations e", "", "declarations decl ; declarations'", "declarations decl ; declarations'", "", "declarations e", "", "", "", "declarations e", "", "", "declarations e", "declarations e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "declarations e"},                                                                                                                                                                                                                                                             // declarations
	{"", "declarations' e", "declarations' e", "", "", "", "declarations' e", "", "declarations' decl ; declarations'", "declarations' decl ; declarations'", "", "declarations' e", "", "", "", "declarations' e", "", "", "declarations' e", "declarations' e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "declarations' e"},                                                                                                                                                                                                                                                   // declarations' // DONE TO HERE
	{"", "", "", "", "", "", "", "", "decl type varlist", "decl type varlist", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                             // decl
	{"", "", "", "", "", "", "", "", "type int", "type double", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                                            // type
	{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "varlist var varlist'"},                                                                                                                                                                                                                                                                                                                                                                                                                           // varlist
	{"", "", "varlist' e", "", "", "", "", "varlist' , varlist", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                                   // varlist'
	{"", "statement_seq statement statement_seq'", "statement_seq statement statement_seq'", "", "", "", "statement_seq statement statement_seq'", "", "", "", "", "statement_seq statement statement_seq'", "", "statement_seq statement statement_seq'", "statement_seq statement statement_seq'", "statement_seq statement statement_seq'", "", "statement_seq statement statement_seq'", "statement_seq statement statement_seq'", "statement_seq statement statement_seq'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "statement_seq statement statement_seq'"},             // statement_seq
	{"", "statement_seq' e", "statement_seq' ; statement_seq", "", "", "", "statement_seq' e", "", "", "", "", "", "", "statement_seq' e", "statement_seq' e", "", "", "statement_seq' e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                 // statement_seq'
	{"", "statement e", "statement e", "", "", "", "statement e", "", "", "", "", "statement if bexpr then statement_seq statement'", "", "statement e", "statement e", "statement while bexpr do statement_seq od", "", "statement e", "statement print expr", "statement return expr", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "statement var = expr"},                                                                                                                                                                                                                       // statement
	{"", "", "", "", "", "", "", "", "", "", "", "", "", "statement' fi", "statement' else statement_seq fi", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                  // statement'
	{"", "", "", "", "expr term expr'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "expr term expr'", "", "", "", "expr term expr'", "expr term expr'", "", "", "", "", "", "", "", "", "", "", "", "expr term expr'"},                                                                                                                                                                                                                                                                                                                                                                    // expr
	{"", "expr' e", "expr' e", "", "", "expr' e", "expr' e", "expr' e", "", "", "", "", "", "expr' e", "expr' e", "", "", "expr' e", "", "", "expr' + term expr'", "expr' - term expr'", "", "", "", "", "", "", "", "", "expr' e", "expr' e", "expr' e", "expr' e", "expr' e", "expr' e", "", "expr' e", ""},                                                                                                                                                                                                                                                                                                  // expr'
	{"", "", "", "", "term factor term'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "term factor term'", "", "", "", "", "", "", "", "", "", "", "", "", "term factor term'"},                                                                                                                                                                                                                                                                                                                                                                                            // term
	{"", "term' e", "term' e", "", "", "term' e", "term' e", "term' e", "", "", "", "", "", "term' e", "term' e", "", "", "term' e", "", "", "term' e", "term' e", "term' * factor term'", "term' / factor term'", "term' % factor term'", "", "", "", "", "", "term' e", "term' e", "term' e", "term' e", "term' e", "term' e", "", "term' e", ""},                                                                                                                                                                                                                                                            // term'
	{"", "", "", "", "factor ( expr )", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "factor INTEGER", "factor DOUBLE", "", "", "", "", "", "", "", "", "", "", "", "factor id factor'"},                                                                                                                                                                                                                                                                                                                                                                                    // factor
	{"", "factor' var'", "factor' var'", "", "factor' ( exprseq )", "factor' var'", "factor' var'", "factor' var'", "", "", "", "", "", "factor' var'", "factor' var'", "", "", "factor' var'", "", "", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "", "", "", "", "", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "factor' var'", "factor' var'", ""},                                                                                                                                                                // factor'
	{"", "", "", "", "exprseq expr exprseq'", "exprseq e", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "exprseq expr exprseq'", "exprseq expr exprseq'", "", "", "", "", "", "", "", "", "", "", "", "exprseq expr exprseq'"},                                                                                                                                                                                                                                                                                                                                                  // exprseq
	{"", "", "", "", "", "exprseq' e", "", "exprseq' , exprseq", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                                   // exprseq'
	{"", "", "", "", "bexpr bterm bexpr'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "bexpr bterm bexpr'", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                           // bexpr
	{"", "", "", "", "", "", "", "", "", "", "", "", "bexpr' e", "", "", "", "bexpr' e", "", "", "", "", "", "", "", "", "", "", "bexpr' or bterm bexpr'", "", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                         // bexpr'
	{"", "", "", "", "bterm bfactor bterm'", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "bterm bfactor bterm'", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                       // bterm
	{"", "", "", "", "", "", "", "", "", "", "", "", "bterm' e", "", "", "", "bterm' e", "", "", "", "", "", "", "", "", "", "", "bterm' e", "bterm' and bfactor bterm'", "", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                              // bterm'
	{"", "", "", "", "bfactor ( expr comp expr )", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "bfactor not bfactor", "", "", "", "", "", "", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                  // bfactor
	{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "comp <", "comp >", "comp ==", "comp <=", "comp >=", "comp <>", "", "", ""},                                                                                                                                                                                                                                                                                                                                                                                                       // comp
	{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "var id var'"},                                                                                                                                                                                                                                                                                                                                                                                                                                    // var
	{"", "var' e", "var' e", "", "", "var' e", "var' e", "var' e", "", "", "var' e", "", "", "var' e", "var' e", "", "", "var' e", "", "", "var' e", "var' e", "var' e", "var' e", "var' e", "", "", "", "", "", "var' e", "var' e", "var' e", "var' e", "var' e", "var' e", "var' [ expr ]", "var' e", ""},                                                                                                                                                                                                                                                                                                    // var'                                                                                                                                                                                                                                                                                                                                                                                                                                       // id
	{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "id IDENTIFIER"},                                                                                                                                                                                                                                                                                                                                                                                                                                  // id
}

var nonTerminalMap = map[string]int{
	"program":        0,
	"fdecls":         2,
	"fdecls'":        2,
	"fdec":           3,
	"params":         4,
	"params'":        5,
	"declarations":   6,
	"declarations'":  7,
	"decl":           8,
	"type":           9,
	"varlist":        10,
	"varlist'":       11,
	"statement_seq":  12,
	"statement_seq'": 13,
	"statement":      14,
	"statement'":     15,
	"expr":           16,
	"expr'":          17,
	"term":           18,
	"term'":          19,
	"factor":         20,
	"factor'":        21,
	"exprseq":        22,
	"exprseq'":       23,
	"bexpr":          24,
	"bexpr'":         25,
	"bterm":          26,
	"bterm'":         27,
	"bfactor":        28,
	"comp":           29,
	"var":            30,
	"var'":           31,
	"id":             32,
}

var terminalMap = map[string]int{
	"$":          0,
	".":          1,
	";":          2,
	"def":        3,
	"(":          4,
	")":          5,
	"fed":        6,
	",":          7,
	"int":        8,
	"double":     9,
	"=":          10,
	"if":         11,
	"then":       12,
	"fi":         13,
	"else":       14,
	"while":      15,
	"do":         16,
	"od":         17,
	"print":      18,
	"return":     19,
	"+":          20,
	"-":          21,
	"*":          22,
	"/":          23,
	"%":          24,
	"INTEGER":    25,
	"DOUBLE":     26,
	"or":         27,
	"and":        28,
	"not":        29,
	"<":          30,
	">":          31,
	"==":         32,
	"<=":         33,
	">=":         34,
	"<>":         35,
	"[":          36,
	"]":          37,
	"IDENTIFIER": 38,
}
