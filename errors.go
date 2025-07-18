package lambdaengine

var mathErrors = map[string]Return{
	// Math
	"incomplete define statement": {"Unable to define variable, incomplete define statement.", true, 100},
	"variable recursion":          {"Unable to define variable, recursive variable assignment.", true, 101},
	"no variable to drop":         {"Unable to drop variable, the variable you are trying to drop does not exist in the current context.", true, 102},
	"incorrect assertion":         {"Unable to define variable, incorrect assertion statement.", true, 103},
	"incomplete drop statement":   {"Unable to drop variable, incomplete drop statement.", true, 104},
	// Lexer
	"multiple decimal splits": {"Unable to parse number, reading multiple decimal splits.", true, 105},
	"number parsing":          {"Unable to parse number, character-conversion faild.", true, 106},
	// Parser
	"missing token":        {"Unable to parse tokens, expecting another token.", true, 107},
	"unclosed parentheses": {"Unable to parse tokens, missing closing parentheses.", true, 108},
	"unopened parentheses": {"Unable to parse tokens, missing opening parentheses.", true, 109},
	"unexpected token":     {"Unable to parse expression, unexpected token.", true, 110},
	//
	"simplify division by zero": {"Unable to simplify calculation, possible devision by zero.", true, 111},
	"undefined variable":        {"Unable to calculate output, undefined variable.", true, 112},
	"division by zero":          {"Unable to calculate output, division by zero.", true, 113},
	"negative sqrt":             {"Unable to calculate output, result has no real solution.", true, 114},
	"unexpected error":          {"Unable to calculate output, unexpected symbole.", true, 115},
}
