package simplifier

import (
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

func isZero(n *shared.Node) bool {
	return n.OperationType == shared.NUMBER && n.Value == 0.0
}

func isNumber(n *shared.Node) bool {
	return n.OperationType == shared.NUMBER
}

// Similar to shared.IsEqual, but it returns true if a and b are factors of each other
// 1, 2 -> true, 2
// a, 2a -> true, 2
// a, a -> true, 1
// a, b -> false, 0
// abcd, abcd -> true, 1
// 2ab, ab -> true, 0.5
// a + b, 2a + 2b -> true, 2
func getMultiple(a, b *shared.Node) (bool, float64) {
	if a.OperationType != b.OperationType {
		if a.OperationType == shared.VARIABLE && b.OperationType == shared.MULTIPLY {
			if ok, factor, variable := getNumFactor(b); ok {
				if variable.Variable == a.Variable {
					return true, factor.Value
				}
			}
		} else if b.OperationType == shared.VARIABLE && a.OperationType == shared.MULTIPLY {
			if ok, factor, variable := getNumFactor(a); ok {
				if variable.Variable == b.Variable {
					return true, 1 / factor.Value
				}
			}
		} else if a.OperationType == shared.MINUS {
			ok, x := getMultiple(a.RNode, b)
			return ok, x * -1
		} else if b.OperationType == shared.MINUS {
			ok, x := getMultiple(a, b.RNode)
			return ok, x * -1
		}
		return false, 0.0

		// Check
	} else if a.OperationType == shared.MULTIPLY {
		factor := 1.0
		// Map for checking if a shared.Node already appeared in the other term.
		alreadySeenB := make(map[*shared.Node]int)

		// Multiply the numbers in term b to the result factor and add all other factors to the alreadySeen map.
		// We shouldn't be able to see the same factor twice, because previously we simplified all duplicate factors to powers?
		for _, bVal := range b.Associative {
			if bVal.OperationType == shared.NUMBER {
				factor = factor * bVal.Value
			} else {
				alreadySeenB[bVal] = 0
			}
		}

		// For each value in the term a, change the result factor or search for the equal in the term b.
		for _, aVal := range a.Associative {
			found := false

			// If current aVal is a Number divide the factor.
			if aVal.OperationType == shared.NUMBER {
				factor = factor / aVal.Value
				continue
			}

			// Else search for the equal factor
			for _, bVal := range b.Associative {
				// Skip if we see a number.
				if bVal.OperationType == shared.NUMBER {
					continue
				}

				// If we have not seen the value already and it is equal to aVal.
				if alreadySeenB[bVal] < 1 {
					if shared.IsEqual(aVal, bVal) {
						// Add it to already seen so it is not checked again later.
						alreadySeenB[bVal]++
						found = true
					}
				}
			}
			// If we did not find a value in term b and it is not a number, we do not have a multple of the other term.
			if !found {
				return false, 0
			}
		}

		// After we have searched for all values of
		for _, bVal := range b.Associative {
			if bVal.OperationType == shared.NUMBER {
				continue
			}
			// If we have not seen a factor of term b in a, return false.
			if alreadySeenB[bVal] < 1 {
				return false, 0
			}
		}
		// Else return true and the factor.
		return true, factor
	} else if a.OperationType == shared.PLUS {
		factor := 1.0
		isFactorDefined := false
		used := make(map[*shared.Node]bool)

		for _, x := range b.Associative {
			if x.OperationType == shared.NUMBER {
				factor = factor * x.Value
			}
		}

		for _, x := range a.Associative {
			if x.OperationType == shared.NUMBER {
				factor = factor / x.Value
				continue
			}

			contains := false
			for _, y := range b.Associative {

				if y.OperationType == shared.NUMBER {
					continue
				}

				if _, ok := used[y]; ok {
					continue
				}

				if ok, fact := getMultiple(x, y); ok {
					if isFactorDefined && factor == fact {
						contains = true
						used[y] = true
						break
					} else if !isFactorDefined {
						isFactorDefined = true
						factor = fact
						contains = true
						used[y] = true
						break
					} else {
						return false, 0
					}
				}
			}
			if !contains {
				return false, 0.0
			}
		}
		return true, factor

	} else if a.OperationType == shared.NUMBER {
		return true, a.Value / b.Value
	}
	return false, 0.0
}

// Returns wether there is a factor of a variable, and if so than it also returns the factor and the variable.
// -> isAFactor, factor, variable
func getNumFactor(node *shared.Node) (bool, *shared.Node, *shared.Node) {
	if node.OperationType == shared.MULTIPLY {
		if len(node.Associative) == 2 {
			if node.Associative[0].OperationType == shared.NUMBER && node.Associative[1].OperationType == shared.VARIABLE {
				return true, node.Associative[0], node.Associative[1]
			} else if node.Associative[0].OperationType == shared.VARIABLE && node.Associative[1].OperationType == shared.NUMBER {
				return true, node.Associative[1], node.Associative[0]
			}
		}
	}
	return false, nil, nil
}

func removeFromNodeArray(a []*shared.Node, i int) []*shared.Node {
	appended := []*shared.Node{}

	for j, val := range a {
		if j == i {

		} else {
			appended = append(appended, val)
		}
	}
	return appended
}

// Checks if a node is either shared.PLUS or shared.MULTIPLY (Cascadable Operation)
func isEndNode(node *shared.Node) bool {
	switch node.OperationType {
	case shared.PLUS, shared.MULTIPLY:
		return false
	default:
		return true
	}
}

// Apply Multiplycation
// a * b = ab
// a * (a + b) = a * a + a * b
func multiplyNodes(x *shared.Node, y *shared.Node) *shared.Node {

	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("(Simplifier - 275:6 - multiplyNodes) {{Debug:}}::cyan|bold Multiplying ")
		cfmt.Printf("%s", shared.PrintATree(x))
		cfmt.Printf(" and ")
		cfmt.Printf("%s", shared.PrintATree(y))
		cfmt.Printf(" to ")
	}

	if x.OperationType == shared.NUMBER && x.Value == 1 {
		if shared.Conf.Options["show_debug_process"] {
			cfmt.Printf("%s", shared.PrintATree(y))
			cfmt.Printf("\n")
		}
		return y
	} else if y.OperationType == shared.NUMBER && y.Value == 1 {
		if shared.Conf.Options["show_debug_process"] {
			cfmt.Printf("%s", shared.PrintATree(x))
			cfmt.Printf("\n")
		}
		return x
	}

	res := &shared.Node{
		OperationType: shared.PLUS,
		Value:         0.0,
		Variable:      "",
		LNode:         nil,
		RNode:         nil,
		Associative:   []*shared.Node{}}

	// Both are at the end of operation
	if isEndNode(x) && isEndNode(y) {
		res = &shared.Node{
			OperationType: shared.MULTIPLY,
			Value:         0.0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   []*shared.Node{x, y},
		}

		// x is added into the multiply operation of y
	} else if y.OperationType == shared.MULTIPLY && isEndNode(x) {
		y.Associative = append(y.Associative, x)
		res = y

		// y is added into the multiply operation of x
	} else if x.OperationType == shared.MULTIPLY && isEndNode(y) {
		x.Associative = append(x.Associative, y)
		res = x

		// x is multiplied by every number in the y operation
	} else if y.OperationType == shared.PLUS && isEndNode(x) {
		for _, val := range y.Associative {
			res.Associative = append(res.Associative, &shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{x, val}})
		}

		// x is multiplied by every number in the y operation
	} else if x.OperationType == shared.PLUS && isEndNode(y) {
		for _, val := range x.Associative {
			res.Associative = append(res.Associative, &shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{y, val}})
		}

	} else if x.OperationType == shared.PLUS && y.OperationType == shared.MULTIPLY {
		for _, val := range x.Associative {
			a := shared.Clone(y)
			a.Associative = append(a.Associative, val)
			res.Associative = append(res.Associative, a)
		}
	} else if x.OperationType == shared.MULTIPLY && y.OperationType == shared.PLUS {
		for _, val := range y.Associative {
			a := shared.Clone(x)
			a.Associative = append(a.Associative, val)
			res.Associative = append(res.Associative, a)
		}

		// If both operations are not an end node, every value is multiplied.
	} else {
		for _, xVal := range x.Associative {
			for _, yVal := range y.Associative {
				res.Associative = append(res.Associative, &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         nil,
					RNode:         nil,
					Associative:   []*shared.Node{xVal, yVal}})
			}
		}
	}
	if shared.Conf.Options["show_debug_process"] {
		cfmt.Printf("%s", shared.PrintATree(res))
		cfmt.Printf("\n")
	}
	return res
}
