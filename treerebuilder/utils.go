package treerebuilder

import (
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

// Goes through all parameters in the function and replaces them with the value.
func replaceFunctionWithValue(parameters []*shared.Node, function *shared.Function) *shared.Node {
	var walk func(node *shared.Node) *shared.Node
	walk = func(node *shared.Node) *shared.Node {
		switch node.OperationType {
		case shared.NUMBER:
			return node
		case shared.MULTIPLY, shared.PLUS:
			newAssociative := []*shared.Node{}
			for _, subnode := range node.Associative {
				newAssociative = append(newAssociative, walk(subnode))
			}
			return &shared.Node{
				OperationType: node.OperationType,
				Value:         0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   newAssociative,
			}
		case shared.POWER:
			return &shared.Node{
				OperationType: node.OperationType,
				Value:         0,
				Variable:      "",
				LNode:         walk(node.LNode),
				RNode:         walk(node.RNode),
				Associative:   nil,
			}
		case shared.VARIABLE:
			for i, val := range function.Parameters {
				if shared.IsEqual(val, node) {
					return parameters[i]
				}
			}
		default:
			cfmt.Printf("Unexpected symbol encountered: %s\n", shared.PrintATree(node))
			return nil
		}
		return nil
	}

	result := shared.Clone(function.Equation)
	result = walk(result)
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("Result of replacement: %s\n", shared.PrintATree(result))
	}

	return result
}
