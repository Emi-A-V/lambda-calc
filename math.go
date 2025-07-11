package main

import (
	"errors"
	"strconv"
	"unicode"

	"github.com/i582/cfmt/cmd/cfmt"
)

const (
	NUMBER       = iota
	PLUS         = iota
	MINUS        = iota
	MULTIPLY     = iota
	DIVIDE       = iota
	POWER        = iota
	SQRT         = iota
	LPARENTHESES = iota
	RPARENTHESES = iota
	EQUAL				 = iota
	VARIABLE		 = iota
)

type Token struct {
	tokenType int
	value     float64
	variable  string
}

type Node struct {
	operationType int
	value         float64
	variable 			string
	lNode         *Node
	rNode         *Node
	associative 	[]*Node
}

var variables map[string]Node = make(map[string]Node)

var variableOccurrence []string

func read(cmd string) (string, error){
	i := 0
	str := ""
	for i < len(cmd) && unicode.IsLetter(rune(cmd[i])){
		str += string(cmd[i])
		i += 1
	}
	switch str {
	case "define":
		if i >= len(cmd) - 1 {
			cfmt.Printf("{{Error:}}::bold|red Unable to define variable, incomplete define statement.\n")
			return "", errors.New("incomplete define statement")
		}

		lexed, err := lexer(cmd[i:])
		if err != nil {
			return "", err
		}
		
		if len(lexed) <= 2 {
			cfmt.Printf("{{Error:}}::bold|red Unable to define variable, incomplete define statement.\n")
			return "", errors.New("incomplete define statement")
		}

		if lexed[0].tokenType == VARIABLE && lexed[1].tokenType == EQUAL {
			node, err := parse(lexed[2:])			
			if err != nil {
				return "", err
			}

			variables[lexed[0].variable] = node
			return "Variable defined.", nil
		}
		cfmt.Printf("{{Error:}}::bold|red Unable to define variable, incorrect assertion statement.\n")
		return "", errors.New("incorrect assertion")
	case "drop":
		if i >= len(cmd) - 1 {
			cfmt.Printf("{{Error:}}::bold|red Unable to drop variable, incomplete drop statement.\n")
			return "", errors.New("incomplete drop statement")
		}

		lexed, err := lexer(cmd[i:])
		if err != nil {
			return "", err
		}

		if _, ok := variables[lexed[0].variable]; ok {
			delete(variables, lexed[0].variable)
			return "Variable deleted.", nil
		} else {
			cfmt.Printf("{{Error:}}::bold|red Unable to drop variable, the variable you are trying to drop does not exist in the current context.\n")
			return "", errors.New("no variable to drop")
		}
	case "solve":
		lexed, err := lexer(cmd)
		if err != nil {
			return "", err
		}
		parsed, err := parse(lexed)
		if err != nil {
			return "", err
		}
		
		solve(&parsed)

		return "", nil
	case "list":
		if len(variables) <= 0 {
			cfmt.Println("No variables defined.")
			return "", nil
		}
		// lists all currently stored variables
		for key, val := range variables {
			cfmt.Println("")
			cfmt.Printf("'%s' : ", key)
			printTree(&val)
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
	lexed, err := lexer(cmd)
	if err != nil {
		return 0, err
	}
	parsed, err := parse(lexed)
	if err != nil {
		return 0, err
	}
	
	// Debug
	if config.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold parse result: ")
		printTree(&parsed)
		cfmt.Println("")
	}

	atred := atr(&parsed)
	// Debug
	if config.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold atr result: ")
		printATree(atred)
		cfmt.Println("")
	}


	simplified, err := simplify(atred, NORMAL)
	if err != nil {
		return 0, err
	}
	
	// Debug 
	if config.Options["show_debug_process"] {
		cfmt.Printf("{{Debug:}}::cyan|bold simplified result: ")
		printATree(simplified)
		cfmt.Println("")
	}

	result, err := eval(simplified, false)
	if err != nil {
		return 0, err
	}
	return result, nil
}



func printATree(node *Node) {
	switch node.operationType {
	case NUMBER:
		cfmt.Print(node.value)
	case VARIABLE:
		if val, ok := variables[node.variable]; ok {
			printATree(&val)
		} else {
			cfmt.Print(node.variable)
		}
	case PLUS:
		cfmt.Print("(")
		for i, val := range node.associative {
			printATree(val)
			if i != len(node.associative) - 1 {
				cfmt.Printf("+")
			}
		}
		cfmt.Print(")")
	case MINUS:
		cfmt.Print("(")
		printATree(node.lNode)
		cfmt.Print("-")
		printATree(node.rNode)
		cfmt.Print(")")
	case MULTIPLY:
		cfmt.Print("(")
		for i, val := range node.associative {
			printATree(val)
			if i != len(node.associative) - 1 {
				cfmt.Printf("*")
			}
		}
		cfmt.Print(")")
	case DIVIDE:
		cfmt.Print("(")
		printATree(node.lNode)
		cfmt.Print("/")
		printATree(node.rNode)
		cfmt.Print(")")
	case POWER:
		cfmt.Print("(")
		printATree(node.lNode)
		cfmt.Print("^")
		printATree(node.rNode)
		cfmt.Print(")")
	case SQRT:
		cfmt.Print("(")
		printATree(node.lNode)
		cfmt.Print("sq")
		printATree(node.rNode)
		cfmt.Print(")")
	}
}

func printTree(node *Node) {
	switch node.operationType {
	case NUMBER:
		cfmt.Print(node.value)
	case VARIABLE:
		if val, ok := variables[node.variable]; ok {
			printTree(&val)
		} else {
			cfmt.Print(node.variable)
		}
	case PLUS:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("+")
		printTree(node.rNode)
		cfmt.Print(")")
	case MINUS:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("-")
		printTree(node.rNode)
		cfmt.Print(")")
	case MULTIPLY:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("*")
		printTree(node.rNode)
		cfmt.Print(")")
	case DIVIDE:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("/")
		printTree(node.rNode)
		cfmt.Print(")")
	case POWER:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("^")
		printTree(node.rNode)
		cfmt.Print(")")
	case SQRT:
		cfmt.Print("(")
		printTree(node.lNode)
		cfmt.Print("sq")
		printTree(node.rNode)
		cfmt.Print(")")
	}
}

