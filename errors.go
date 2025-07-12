package lambdaengine

import "errors"

var mathErrors = map[error]Return {
	// Math
	errors.New("incomplete define statement")	: {"Unable to define variable, incomplete define statement.", true, 100},
	errors.New("variable recursion")					:	{"Unable to define variable, recursive variable assignment.", true, 101},
	errors.New("no variable to drop")					: {"Unable to drop variable, the variable you are trying to drop does not exist in the current context.", true, 102},
	errors.New("incorrect assertion") 				: {"Unable to define variable, incorrect assertion statement.", true, 103},
	errors.New("incomplete drop statement") 	: {"Unable to drop variable, incomplete drop statement.", true, 104},
	// Lexer
	errors.New("multiple decimal splits") 		: {"Unable to parse number, reading multiple decimal splits.", true, 105},
	errors.New("number parsing") 							: {"Unable to parse number, character-conversion faild.", true, 106},
	// Parser
	errors.New("missing token") 							: {"Unable to parse tokens, expecting another token.", true, 107},
	errors.New("unclosed parentheses") 				: {"Unable to parse tokens, missing closing parentheses.", true, 108},
	errors.New("unclosed parentheses") 				: {"Unable to parse tokens, missing closing parentheses.", true, 109},
	errors.New("unexpected token") 						: {"Unable to parse expression, unexpected token.", true, 110},
	// 
	errors.New("simplify division by zero") 	: {"Unable to simplify calculation, possible devision by zero.", true, 111},
	errors.New("undefined variable") 					: {"Unable to calculate output, undefined variable.", true, 112},
	errors.New("division by zero") 						: {"Unable to calculate output, division by zero.", true, 113},
	errors.New("negative sqrt") 							: {"Unable to calculate output, result has no real solution.", true, 114},
	errors.New("unexpected error") 						: {"Unable to calculate output, unexpected symbole.", true, 115},
}
