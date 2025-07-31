package treerebuilder

import "lambdacalc/shared"

// The Associative Tree Rebuilder rebuilds the tree so that associative rules are followed,
// i.e.: a + b + c = c + b + a
// After `AssociativeTreeRebuild` is used, all nodes of the type `shared.MULTIPLY` and `shared.PLUS`, use the `associative` variabe to
// show all of its child nodes.

func AssociativeTreeRebuild(node *shared.Node) *shared.Node {
	switch node.OperationType {
	case shared.NUMBER, shared.VARIABLE:
		return node
	case shared.MINUS, shared.PLUS, shared.MULTIPLY, shared.DIVIDE:
		var result []*shared.Node
		var walk func(n *shared.Node)
		var op int

		switch node.OperationType {
		case shared.MINUS:
			op = shared.PLUS
		case shared.DIVIDE:
			op = shared.MULTIPLY
		default:
			op = node.OperationType
		}

		walk = func(n *shared.Node) {
			if n.OperationType == op && n.OperationType != shared.MINUS && n.OperationType != shared.DIVIDE {
				walk(n.LNode)
				walk(n.RNode)
			} else if op == shared.PLUS && n.OperationType == shared.MINUS {
				walk(n.LNode)
				result = append(result, &shared.Node{
					OperationType: shared.MINUS,
					Value:         0.0,
					Variable:      "",
					LNode: &shared.Node{
						OperationType: shared.NUMBER,
						Value:         0.0,
						Variable:      "",
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					},
					RNode:       AssociativeTreeRebuild(n.RNode),
					Associative: nil,
				})
			} else if op == shared.MULTIPLY && n.OperationType == shared.DIVIDE {
				walk(n.LNode)
				result = append(result, &shared.Node{
					OperationType: shared.POWER,
					Value:         0.0,
					Variable:      "",
					LNode:         AssociativeTreeRebuild(n.RNode),
					RNode: &shared.Node{
						OperationType: shared.NUMBER,
						Value:         -1.0,
						Variable:      "",
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					},
					Associative: nil,
				})
			} else {
				result = append(result, AssociativeTreeRebuild(n))
			}
		}
		walk(node)
		return &shared.Node{
			OperationType: op,
			Value:         0.0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   result,
		}
	default:
		return &shared.Node{
			OperationType: node.OperationType,
			Value:         0.0,
			Variable:      "",
			LNode:         AssociativeTreeRebuild(node.LNode),
			RNode:         AssociativeTreeRebuild(node.RNode),
			Associative:   nil,
		}
	}
}
