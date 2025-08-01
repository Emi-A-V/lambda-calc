package main

import (
	"lambdacalc/shared"
	"lambdacalc/utils"

	"lambdacalc/interpreter"
	"lambdacalc/lexer"
	"lambdacalc/parser"
	"lambdacalc/simplifier"
	"lambdacalc/solver"
	"lambdacalc/treerebuilder"

	"errors"
	"strconv"
	"unicode"

	"github.com/i582/cfmt/cmd/cfmt"
)

func read(cmd string) (string, error) {
	i := 0
	str := ""

	// VariableOccurrence := []string{}

	for i < len(cmd) && unicode.IsLetter(rune(cmd[i])) {
		str += string(cmd[i])
		i += 1
	}
	switch str {
	case "define":
		if i >= len(cmd)-1 {
			cfmt.Printf("{{Error:}}::bold|red Unable to define variable, incomplete define statement.\n")
			return "", errors.New("incomplete define statement")
		}

		lexed, err := lexer.LexTokens(cmd[i:])
		if err != nil {
			return "", err
		}

		if len(lexed) <= 2 {
			cfmt.Printf("{{Error:}}::bold|red Unable to define variable, incomplete define statement.\n")
			return "", errors.New("incomplete define statement")
		}

		parsed, err := parser.Parse(lexed)
		if err != nil {
			return "", err
		}

		// Check if all parameters are variables.
		if parsed.OperationType == shared.EQUAL && parsed.LNode.OperationType == shared.FUNCTION {
			for _, val := range parsed.LNode.Associative {
				if val.OperationType != shared.VARIABLE {
					cfmt.Printf("{{Error:}}::bold|red Unable to define function, functions can only be declared with variables as parameters.\n")
					return "", errors.New("unallowed operation in function declaration")
				}
			}
		}
		atr := treerebuilder.AssociativeTreeRebuild(&parsed)
		simplified, err := simplifier.Simplify(atr, simplifier.UNWIND)
		if err != nil {
			return "", err
		}

		if parsed.OperationType == shared.EQUAL && parsed.LNode.OperationType == shared.VARIABLE {
			shared.Variables[lexed[0].Variable] = *simplified.RNode
			return "Variable defined.", nil
		} else if parsed.OperationType == shared.EQUAL && parsed.LNode.OperationType == shared.FUNCTION {
			shared.Functions[lexed[0].Variable] = shared.Function{
				Parameters: simplified.LNode.Associative,
				Equation:   simplified.RNode,
			}
			return "Function defined.", nil
		} else {
			cfmt.Printf("{{Error:}}::bold|red Unable to define variable or function, incorrect assertion statement.\n")
			return "", errors.New("incorrect assertion")
		}
	case "drop":
		if i >= len(cmd)-1 {
			cfmt.Printf("{{Error:}}::bold|red Unable to drop variable, incomplete drop statement.\n")
			return "", errors.New("incomplete drop statement")
		}

		lexed, err := lexer.LexTokens(cmd[i:])
		if err != nil {
			return "", err
		}

		if _, ok := shared.Variables[lexed[0].Variable]; ok {
			delete(shared.Variables, lexed[0].Variable)
			return "Variable deleted.", nil
		} else if _, ok := shared.Functions[lexed[0].Variable]; ok {
			delete(shared.Functions, lexed[0].Variable)
			return "Variable deleted.", nil
		} else {
			cfmt.Printf("{{Error:}}::bold|red Unable to drop variable, the variable you are trying to drop does not exist in the current context.\n")
			return "", errors.New("no variable to drop")
		}
	case "solve":
		lexed, err := lexer.LexTokens(cmd)
		if err != nil {
			return "", err
		}
		parsed, err := parser.Parse(lexed)
		if err != nil {
			return "", err
		}

		solver.Solve(&parsed)

		return "", nil
	case "list":
		if len(shared.Variables) <= 0 {
			cfmt.Println("No shared.Variables defined.")
			return "", nil
		}
		// lists all currently stored shared.Variables
		for key, val := range shared.Variables {
			cfmt.Println("")
			cfmt.Printf("'%s' : ", key)
			utils.PrintTree(&val)
		}
		cfmt.Println("")
		return "", nil
	default:
		num, err := calc(cmd)
		cfmt.Println("")
		numStr := ""
		if err != nil {
			numStr = ""
		} else {
			numStr = strconv.FormatFloat(num, 'f', -1, 64)
		}
		return numStr, err
	}
}

func calc(cmd string) (float64, error) {
	lexed, err := lexer.LexTokens(cmd)
	if err != nil {
		return 0, err
	}
	parsed, err := parser.Parse(lexed)
	if err != nil {
		return 0, err
	}

	// Debug
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold parse result: ")
		utils.PrintTree(&parsed)
		cfmt.Println("")
	}

	atred := treerebuilder.AssociativeTreeRebuild(&parsed)
	// Debug
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold atr result: ")
		cfmt.Printf("%s", utils.PrintATree(atred))
		cfmt.Println("")
	}

	unwound, err := simplifier.Simplify(atred, simplifier.UNWIND)
	if err != nil {
		return 0, err
	}

	// Debug
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold Unwound result: ")
		cfmt.Printf("%s", utils.PrintATree(unwound))
		cfmt.Println("")
	}

	rewound, err := simplifier.Simplify(unwound, simplifier.REWIND)
	if err != nil {
		return 0, err
	}

	// Debug
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold Rewound result: ")
		cfmt.Printf("%s", utils.PrintATree(rewound))
		cfmt.Println("")
	}

	result, err := interpreter.Evaluate(rewound, false)
	if err != nil {
		return 0, err
	}
	return result, nil
}
