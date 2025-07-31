package utils

import (
	"lambdacalc/shared"
	"strconv"
)

func PrintATree(node *shared.Node) string {
	str := ""
	switch node.OperationType {
	case shared.NUMBER:
		str += strconv.FormatFloat(node.Value, 'f', -1, 64)
	case shared.VARIABLE:
		if val, ok := shared.Variables[node.Variable]; ok {
			str += PrintATree(&val)
		} else {
			str += node.Variable
		}
	case shared.PLUS:
		str += "("
		for i, val := range node.Associative {
			str += PrintATree(val)
			if i != len(node.Associative)-1 {
				str += "+"
			}
		}
		str += ")"
	case shared.MINUS:
		str += "("
		str += PrintATree(node.LNode)
		str += "-"
		str += PrintATree(node.RNode)
		str += ")"
	case shared.MULTIPLY:
		str += "("
		for i, val := range node.Associative {
			str += PrintATree(val)
			if i != len(node.Associative)-1 {
				str += "*"
			}
		}
		str += ")"
	case shared.DIVIDE:
		str += "("
		str += PrintATree(node.LNode)
		str += "/"
		str += PrintATree(node.RNode)
		str += ")"
	case shared.POWER:
		str += "("
		str += PrintATree(node.LNode)
		str += "^"
		str += PrintATree(node.RNode)
		str += ")"
	case shared.SQRT:
		str += "("
		str += PrintATree(node.LNode)
		str += "sq"
		str += PrintATree(node.RNode)
		str += ")"
	}
	return str
}

func PrintTree(node *shared.Node) string {
	str := ""
	switch node.OperationType {
	case shared.NUMBER:
		str += strconv.FormatFloat(node.Value, 'f', -1, 64)
	case shared.VARIABLE:
		if val, ok := shared.Variables[node.Variable]; ok {
			str += PrintTree(&val)
		} else {
			str += node.Variable
		}
	case shared.PLUS:
		str += "("
		str += PrintTree(node.LNode)
		str += "+"
		str += PrintTree(node.RNode)
		str += ")"
	case shared.MINUS:
		str += "("
		str += PrintTree(node.LNode)
		str += "-"
		str += PrintTree(node.RNode)
		str += ")"
	case shared.MULTIPLY:
		str += "("
		str += PrintTree(node.LNode)
		str += "*"
		str += PrintTree(node.RNode)
		str += ")"
	case shared.DIVIDE:
		str += "("
		str += PrintTree(node.LNode)
		str += "/"
		str += PrintTree(node.RNode)
		str += ")"
	case shared.POWER:
		str += "("
		str += PrintTree(node.LNode)
		str += "^"
		str += PrintTree(node.RNode)
		str += ")"
	case shared.SQRT:
		str += "("
		str += PrintTree(node.LNode)
		str += "sq"
		str += PrintTree(node.RNode)
		str += ")"
	}
	return str
}
