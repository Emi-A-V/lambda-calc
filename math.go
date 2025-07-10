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
}

var variables map[string]Node = make(map[string]Node)

func read(cmd string) (string, error){
	i := 0
	str := ""
	for i < len(cmd) && unicode.IsLetter(rune(cmd[i])){
		str += string(cmd[i])
		i += 1
	}
	switch str {
	case "define":
		lexed, err := lexer(cmd[i:])
		if err != nil {
			return "", err
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
		printTree(&parsed)
		cfmt.Println("")
	}

	simplified, err := simplify(&parsed, NORMAL)
	if err != nil {
		return 0, err
	}
	
	// Debug 
	if config.Options["show_debug_process"] {
		printTree(simplified)
		cfmt.Println("")
	}

	result, err := eval(simplified, false)
	if err != nil {
		return 0, err
	}
	return result, nil
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

