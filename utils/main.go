package utils

import (
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

func PrintATree(node *shared.Node) {
	switch node.OperationType {
	case shared.NUMBER:
		cfmt.Print(node.Value)
	case shared.VARIABLE:
		if val, ok := shared.Variables[node.Variable]; ok {
			PrintATree(&val)
		} else {
			cfmt.Print(node.Variable)
		}
	case shared.PLUS:
		cfmt.Print("(")
		for i, val := range node.Associative {
			PrintATree(val)
			if i != len(node.Associative)-1 {
				cfmt.Printf("+")
			}
		}
		cfmt.Print(")")
	case shared.MINUS:
		cfmt.Print("(")
		PrintATree(node.LNode)
		cfmt.Print("-")
		PrintATree(node.RNode)
		cfmt.Print(")")
	case shared.MULTIPLY:
		cfmt.Print("(")
		for i, val := range node.Associative {
			PrintATree(val)
			if i != len(node.Associative)-1 {
				cfmt.Printf("*")
			}
		}
		cfmt.Print(")")
	case shared.DIVIDE:
		cfmt.Print("(")
		PrintATree(node.LNode)
		cfmt.Print("/")
		PrintATree(node.RNode)
		cfmt.Print(")")
	case shared.POWER:
		cfmt.Print("(")
		PrintATree(node.LNode)
		cfmt.Print("^")
		PrintATree(node.RNode)
		cfmt.Print(")")
	case shared.SQRT:
		cfmt.Print("(")
		PrintATree(node.LNode)
		cfmt.Print("sq")
		PrintATree(node.RNode)
		cfmt.Print(")")
	}
}

func PrintTree(node *shared.Node) {
	switch node.OperationType {
	case shared.NUMBER:
		cfmt.Print(node.Value)
	case shared.VARIABLE:
		if val, ok := shared.Variables[node.Variable]; ok {
			PrintTree(&val)
		} else {
			cfmt.Print(node.Variable)
		}
	case shared.PLUS:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("+")
		PrintTree(node.RNode)
		cfmt.Print(")")
	case shared.MINUS:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("-")
		PrintTree(node.RNode)
		cfmt.Print(")")
	case shared.MULTIPLY:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("*")
		PrintTree(node.RNode)
		cfmt.Print(")")
	case shared.DIVIDE:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("/")
		PrintTree(node.RNode)
		cfmt.Print(")")
	case shared.POWER:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("^")
		PrintTree(node.RNode)
		cfmt.Print(")")
	case shared.SQRT:
		cfmt.Print("(")
		PrintTree(node.LNode)
		cfmt.Print("sq")
		PrintTree(node.RNode)
		cfmt.Print(")")
	}
}
