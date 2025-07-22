package lambdaengine

import (
	"fmt"
	"slices"
	"strconv"
	"unicode"
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
	EQUAL        = iota
	VARIABLE     = iota
)

type Token struct {
	tokenType int
	value     float64
	variable  string
}

type Node struct {
	operationType int
	value         float64
	variable      string
	lNode         *Node
	rNode         *Node
	associative   []*Node
}

type Return struct {
	Str   string
	Err   bool
	ErrID int
}

type Var struct {
	Var string
	Dep []string
	Equ string
}

var variables map[string]Node = make(map[string]Node)

var variableOccurrence []string

func Input(cmd string) (Return, string) {
	i := 0
	str := ""

	variableOccurrence := []string{}

	for i < len(cmd) && unicode.IsLetter(rune(cmd[i])) {
		str += string(cmd[i])
		i += 1
	}
	switch str {
	case "define":
		if i >= len(cmd)-1 {
			return mathErrors["incomplete define statement"], ""
		}

		lexed, err := lexer(cmd[i:])
		if err != nil {
			return mathErrors[err.Error()], ""
		}

		if len(lexed) <= 2 {
			return mathErrors["incomplete define statement"], ""
		}

		if lexed[0].tokenType == VARIABLE && lexed[1].tokenType == EQUAL {

			if len(variableOccurrence) >= 2 {
				if slices.Contains(variableOccurrence[1:], variableOccurrence[0]) {
					return mathErrors["variable recursion"], ""
				}
			}

			node, err := parse(lexed[2:])
			if err != nil {
				return mathErrors[err.Error()], ""
			}

			variables[lexed[0].variable] = node
			eventVariableDefinedCallback(Var{lexed[0].variable, []string{}, printTree(&node)})
			return Return{"Variable defined.", false, 201}, ""
		}
		return mathErrors["incorrect assertion"], ""
	case "drop":
		if i >= len(cmd)-1 {
			return mathErrors["incomplete drop statement"], ""
		}

		lexed, err := lexer(cmd[i:])
		if err != nil {
			return mathErrors[err.Error()], ""
		}

		if val, ok := variables[lexed[0].variable]; ok {
			eventVariableDroppedCallback(Var{lexed[0].variable, []string{}, printTree(&val)})
			delete(variables, lexed[0].variable)
			return Return{"Variable deleted.", false, 202}, ""
		} else {
			return mathErrors["no variable to drop"], ""
		}
	case "solve":
		lexed, err := lexer(cmd)
		if err != nil {
			return mathErrors[err.Error()], ""
		}
		parsed, err := parse(lexed)
		if err != nil {
			return mathErrors[err.Error()], ""
		}

		solve(&parsed)

		return Return{"", false, 200}, ""
	case "list":
		if len(variables) <= 0 {
			return Return{"No variables defined yet.", false, 200}, ""
		}

		str := ""
		// lists all currently stored variables
		for key, val := range variables {
			str += fmt.Sprintf("'%s' : ", key)
			str += printATree(&val)
			str += "\n"
		}
		return Return{"", false, 203}, str
	default:
		num, err := calc(cmd)
		if err != nil {
			return mathErrors[err.Error()], num
		} else {
			return Return{"", false, 200}, num
		}
	}
}

func calc(cmd string) (string, error) {
	lexed, err := lexer(cmd)
	if err != nil {
		return "", err
	}
	parsed, err := parse(lexed)
	if err != nil {
		return "", err
	}

	atred := atr(&parsed)

	simplified, err := simplify(atred, UNWIND)
	if err != nil {
		return printATree(atred), err
	}
	result, err := eval(simplified, false)
	if err != nil {
		return printATree(simplified), err
	}
	return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func printATree(node *Node) string {
	str := ""
	switch node.operationType {
	case NUMBER:
		str += strconv.FormatFloat(node.value, 'f', -1, 64)
	case VARIABLE:
		if val, ok := variables[node.variable]; ok {
			str += printATree(&val)
		} else {
			str += node.variable
		}
	case PLUS:
		str += "("
		for i, val := range node.associative {
			str += printATree(val)
			if i != len(node.associative)-1 {
				str += "+"
			}
		}
		str += ")"
	case MINUS:
		str += "("
		str += printATree(node.lNode)
		str += "-"
		str += printATree(node.rNode)
		str += ")"
	case MULTIPLY:
		str += "("
		for i, val := range node.associative {
			str += printATree(val)
			if i != len(node.associative)-1 {
				str += "*"
			}
		}
		str += ")"
	case DIVIDE:
		str += "("
		str += printATree(node.lNode)
		str += "/"
		str += printATree(node.rNode)
		str += ")"
	case POWER:
		str += "("
		str += printATree(node.lNode)
		str += "^"
		str += printATree(node.rNode)
		str += ")"
	case SQRT:
		str += "("
		str += printATree(node.lNode)
		str += "sq"
		str += printATree(node.rNode)
		str += ")"
	}
	return str
}

func printTree(node *Node) string {
	str := ""
	switch node.operationType {
	case NUMBER:
		str += strconv.FormatFloat(node.value, 'f', -1, 64)
	case VARIABLE:
		if val, ok := variables[node.variable]; ok {
			str += printTree(&val)
		} else {
			str += node.variable
		}
	case PLUS:
		str += "("
		str += printTree(node.lNode)
		str += "+"
		str += printTree(node.rNode)
		str += ")"
	case MINUS:
		str += "("
		str += printTree(node.lNode)
		str += "-"
		str += printTree(node.rNode)
		str += ")"
	case MULTIPLY:
		str += "("
		str += printTree(node.lNode)
		str += "*"
		str += printTree(node.rNode)
		str += ")"
	case DIVIDE:
		str += "("
		str += printTree(node.lNode)
		str += "/"
		str += printTree(node.rNode)
		str += ")"
	case POWER:
		str += "("
		str += printTree(node.lNode)
		str += "^"
		str += printTree(node.rNode)
		str += ")"
	case SQRT:
		str += "("
		str += printTree(node.lNode)
		str += "sq"
		str += printTree(node.rNode)
		str += ")"
	}
	return str
}
