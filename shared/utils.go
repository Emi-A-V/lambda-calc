package shared

import (
	"strconv"

	"github.com/i582/cfmt/cmd/cfmt"
)

func ZeroNode() *Node {
	return &Node{
		OperationType: NUMBER,
		Value:         0.0,
		Variable:      "",
		LNode:         nil,
		RNode:         nil,
		Associative:   nil,
	}
}

func IsEqual(a, b *Node) bool {
	if a.OperationType != b.OperationType {
		return false
	} else if a.OperationType == NUMBER {
		return a.Value == b.Value
	} else if a.OperationType == DIVIDE {
		return IsEqual(a.LNode, b.LNode) && IsEqual(a.RNode, b.RNode)
	} else if a.OperationType == VARIABLE {
		return a.Variable == b.Variable
	} else if a.OperationType == MULTIPLY || a.OperationType == PLUS {
		return containSameNodes(a.Associative, b.Associative)
	}
	return true
}

// Frequenzy Comparision of Node arrays.
// TODO: Pointers might not be the same for two of the same items...
func containSameNodes(a []*Node, b []*Node) bool {
	if len(a) != len(b) {
		return false
	}

	used := make(map[*Node]bool)

	for _, x := range a {
		contains := false
		for _, y := range b {

			// Skip if already used
			if _, ok := used[y]; ok {
				continue
			}

			// Check if equal
			if IsEqual(x, y) {
				contains = true
				used[y] = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

// TODO: WTF this is wrong!
func Clone(n *Node) *Node {
	if n == nil {
		return nil
	}

	copy := *n

	copy.LNode = Clone(n.LNode)
	copy.RNode = Clone(n.RNode)

	copy.Associative = []*Node{}
	for _, val := range n.Associative {
		copy.Associative = append(copy.Associative, Clone(val))
	}

	return &copy
}

func PrintATree(node *Node) string {
	if node == nil {
		cfmt.Printf("Encountered unexpected null pointer reference.\n")
		return ""
	}

	str := ""
	switch node.OperationType {
	case NUMBER:
		str += strconv.FormatFloat(node.Value, 'f', -1, 64)
	case VARIABLE:
		if val, ok := Variables[node.Variable]; ok {
			str += PrintATree(&val)
		} else {
			str += node.Variable
		}
	case PLUS:
		str += "("
		for i, val := range node.Associative {
			str += PrintATree(val)
			if i != len(node.Associative)-1 {
				str += "+"
			}
		}
		str += ")"
	case MINUS:
		str += "("
		str += PrintATree(node.LNode)
		str += "-"
		str += PrintATree(node.RNode)
		str += ")"
	case MULTIPLY:
		str += "("
		for i, val := range node.Associative {
			str += PrintATree(val)
			if i != len(node.Associative)-1 {
				str += "*"
			}
		}
		str += ")"
	case DIVIDE:
		str += "("
		str += PrintATree(node.LNode)
		str += "/"
		str += PrintATree(node.RNode)
		str += ")"
	case POWER:
		str += "("
		str += PrintATree(node.LNode)
		str += "^"
		str += PrintATree(node.RNode)
		str += ")"
	case SQRT:
		str += "("
		str += PrintATree(node.LNode)
		str += "sq"
		str += PrintATree(node.RNode)
		str += ")"
	case COMMA:
		str += ","
	case FUNCTION:
		str += node.Variable
		str += "("
		for i, val := range node.Associative {
			str += PrintATree(val)
			if i != len(node.Associative)-1 {
				str += ", "
			}
		}
		str += ")"
	}
	return str
}

func PrintTree(node *Node) string {
	str := ""
	switch node.OperationType {
	case NUMBER:
		str += strconv.FormatFloat(node.Value, 'f', -1, 64)
	case VARIABLE:
		if val, ok := Variables[node.Variable]; ok {
			str += PrintTree(&val)
		} else {
			str += node.Variable
		}
	case PLUS:
		str += "("
		str += PrintTree(node.LNode)
		str += "+"
		str += PrintTree(node.RNode)
		str += ")"
	case MINUS:
		str += "("
		str += PrintTree(node.LNode)
		str += "-"
		str += PrintTree(node.RNode)
		str += ")"
	case MULTIPLY:
		str += "("
		str += PrintTree(node.LNode)
		str += "*"
		str += PrintTree(node.RNode)
		str += ")"
	case DIVIDE:
		str += "("
		str += PrintTree(node.LNode)
		str += "/"
		str += PrintTree(node.RNode)
		str += ")"
	case POWER:
		str += "("
		str += PrintTree(node.LNode)
		str += "^"
		str += PrintTree(node.RNode)
		str += ")"
	case SQRT:
		str += "("
		str += PrintTree(node.LNode)
		str += "sq"
		str += PrintTree(node.RNode)
		str += ")"
	case COMMA:
		str += ","
	case FUNCTION:
		str += node.Variable
		str += "("
		for i, val := range node.Associative {
			str += PrintTree(val)
			if i != len(node.Associative)-1 {
				str += ", "
			}
		}
		str += ")"
	}
	return str
}
