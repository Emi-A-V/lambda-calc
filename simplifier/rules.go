package simplifier

import (
	"errors"
	"lambdacalc/interpreter"
	"lambdacalc/shared"
	"lambdacalc/utils"
	"slices"

	"github.com/i582/cfmt/cmd/cfmt"
)

// x + 0 = x
func simplifyAddZero(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.PLUS {
		changed := false
		for i, val := range node.Associative {
			if isZero(val) {
				changed = true
				node.Associative = removeFromNodeArray(node.Associative, i)
				i--
			}
		}
		if changed {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// x - 0 = x
func simplifySubZero(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MINUS {
		if isZero(node.RNode) {
			return node.RNode, true, nil
		}
	}
	return nil, false, nil
}

// x * 0 = 0
func simplifyMultZero(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MULTIPLY {
		if slices.ContainsFunc(node.Associative, isZero) {
			return &shared.Node{
				OperationType: shared.NUMBER,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, true, nil
		}
	}
	return nil, false, nil
}

// x * 1 = x
func simplifyMultOne(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MULTIPLY {
		changed := false
		for i := 0; i < len(node.Associative); i++ {
			val := node.Associative[i]
			if val.OperationType == shared.NUMBER && val.Value == 1.0 {
				changed = true
				node.Associative = removeFromNodeArray(node.Associative, i)
				i--
			}
		}
		if changed {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// x / 1 = x
func simplifyDivOne(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.DIVIDE {
		if isNumber(node.RNode) && node.RNode.Value == 1 {
			return clone(node.LNode), true, nil
		}
	}
	return nil, false, nil
}

// 0 / x = 0 (x != 0)
// LATER: Maybe add assumptions
func simplifyZeroDiv(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.DIVIDE {
		if isNumber(node.LNode) && node.LNode.Value == 0 {
			if val, err := interpreter.Evaluate(node.RNode, true); err == nil && val != 0 {
				return &shared.Node{
					OperationType: shared.NUMBER,
					Value:         0.0,
					Variable:      "",
					LNode:         nil,
					RNode:         nil,
					Associative:   nil,
				}, true, nil
			} else if val == 0 {
				cfmt.Printf("{{Error:}}::red|bold Unable to simplify calculation, possible devision by zero.\n")
				return nil, false, errors.New("divide by 0")
			}
		}
	}
	return nil, false, nil
}

// x / x = 1
func simplifyDivSelf(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.DIVIDE {
		if isEqual(node.RNode, node.LNode) {
			return &shared.Node{
				OperationType: shared.NUMBER,
				Value:         1.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, true, nil
		}
	}
	return nil, false, nil
}

// (+(a*b)) = (a*b)
// After collection of addition, clean up
func simplifySingleAdd(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.PLUS || node.OperationType == shared.MULTIPLY {
		if len(node.Associative) == 1 {
			return node.Associative[0], true, nil
		}
	}
	return nil, false, nil
}

// Collect all terms in `shared.PLUS` operations.
// 2 + 4 + a + b + a = 6 + 2a + b
func simplifyAddCollect(n *shared.Node) (*shared.Node, bool, error) {
	if n.OperationType == shared.PLUS {
		node := clone(n)
		changed := false

		nNumOp := 0
		nVarOp := 0

		result := 0.0
		varMap := make(map[string]float64)

		for i := 0; i < len(node.Associative); i++ {
			val := node.Associative[i]
			switch val.OperationType {
			case shared.NUMBER:
				result += val.Value
				node.Associative = removeFromNodeArray(node.Associative, i)
				nNumOp++
				i--
			case shared.VARIABLE:
				varMap[val.Variable]++
				node.Associative = removeFromNodeArray(node.Associative, i)
				if varMap[val.Variable] != 1 {
					nVarOp++
				}
				i--
			case shared.MINUS:
				switch val.RNode.OperationType {
				case shared.VARIABLE:
					varMap[val.RNode.Variable]--
					node.Associative = removeFromNodeArray(node.Associative, i)
					if varMap[val.Variable] != -1 {
						nVarOp += 2
					}
					i--
				}
			case shared.MULTIPLY:
				num := 1.0
				newVal := clone(val)

				// Knowing that if we found a previous multiple of this term, it should already be simplified.
				// Searching for next multiple of the current term.
				for y := i + 1; y < len(node.Associative); y++ {
					// If we find a term that is a multiple..
					if ok, fact := getMultiple(val, node.Associative[y]); ok {
						if shared.Conf.Options["show_debug_process"] {
							cfmt.Printf("(Simplifier - 648:8 - simplifyAddCollect) {{Debug:}}::cyan|bold Found factor in addends: ")
							utils.PrintATree(val)
							cfmt.Printf(" / ")
							utils.PrintATree(node.Associative[y])
							cfmt.Printf(" = ")
							cfmt.Printf("%v\n", fact)
						}
						num += fact
						node.Associative = removeFromNodeArray(node.Associative, y)
						y--
						nVarOp += 2
						changed = true
					}
				}
				if num != 1 {
					newVal.Associative = append(newVal.Associative, &shared.Node{
						OperationType: shared.NUMBER,
						Value:         num,
						Variable:      "",
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					})
				}
				node.Associative[i] = newVal
			}
		}
		if result != 0.0 {
			node.Associative = append(node.Associative, &shared.Node{
				OperationType: shared.NUMBER,
				Value:         result,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				switch fact {
				case 0:
					changed = true
				case 1:
					node.Associative = append(node.Associative, &shared.Node{
						OperationType: shared.VARIABLE,
						Value:         0.0,
						Variable:      key,
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					})
					changed = true
				default:
					mult := &shared.Node{
						OperationType: shared.MULTIPLY,
						Value:         0.0,
						Variable:      "",
						LNode:         nil,
						RNode:         nil,
						Associative: []*shared.Node{
							{
								OperationType: shared.VARIABLE,
								Value:         0.0,
								Variable:      key,
								LNode:         nil,
								RNode:         nil,
								Associative:   nil,
							},
							{
								OperationType: shared.NUMBER,
								Value:         float64(fact),
								Variable:      "",
								LNode:         nil,
								RNode:         nil,
								Associative:   nil,
							},
						},
					}
					node.Associative = append(node.Associative, mult)
					changed = true
				}
			}
		}
		if changed && (nVarOp >= 2 || nNumOp >= 2) {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// Collect all terms in `shared.MULTIPLY` operations.
// a * a * b * 2 * 5 = 10 * b * a^2
func simplifyMultCollect(n *shared.Node) (*shared.Node, bool, error) {
	if n.OperationType == shared.MULTIPLY {
		node := clone(n)
		changed := false

		nNumOp := 0
		nVarOp := 0

		result := 1.0
		varMap := make(map[string]float64)

		for i := 0; i < len(node.Associative); i++ {
			val := node.Associative[i]
			switch val.OperationType {
			case shared.NUMBER:
				result *= val.Value
				node.Associative = removeFromNodeArray(node.Associative, i)
				nNumOp++
				i--
			case shared.VARIABLE:
				varMap[val.Variable]++
				node.Associative = removeFromNodeArray(node.Associative, i)
				if varMap[val.Variable] != 1 {
					nVarOp++
				}
				i--
			case shared.MINUS:
				if val.RNode.OperationType == shared.VARIABLE {
					varMap[val.RNode.Variable]++
					result *= -1.0
					node.Associative = removeFromNodeArray(node.Associative, i)
					if varMap[val.RNode.Variable] != 1 {
						nVarOp++
					}
					nNumOp += 2
					i--
				}
			case shared.POWER:
				if val.LNode.OperationType == shared.VARIABLE && val.RNode.OperationType == shared.NUMBER {
					varMap[val.LNode.Variable] += val.RNode.Value
					node.Associative = removeFromNodeArray(node.Associative, i)
					if varMap[val.LNode.Variable] != val.RNode.Value {
						nVarOp++
					}
					i--

				} else if val.LNode.OperationType == shared.NUMBER && val.RNode.OperationType == shared.NUMBER && val.RNode.Value == -1 {
					result = result / val.LNode.Value
					node.Associative = removeFromNodeArray(node.Associative, i)
					nNumOp++
					i--
				}
			}
		}
		if nNumOp >= 2 && nVarOp >= 2 && result == 1.0 {
			changed = true
		} else if result != 1 {
			node.Associative = append(node.Associative, &shared.Node{
				OperationType: shared.NUMBER,
				Value:         result,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				if fact == 1 {
					node.Associative = append(node.Associative, &shared.Node{
						OperationType: shared.VARIABLE,
						Value:         0.0,
						Variable:      key,
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					})
					changed = true
					// If the Variables multiply together to x^0, replace them with 1
				} else if fact == 0 {
					node.Associative = append(node.Associative, &shared.Node{
						OperationType: shared.NUMBER,
						Value:         1.0,
						Variable:      "",
						LNode:         nil,
						RNode:         nil,
						Associative:   nil,
					})
					changed = true
				} else {
					mult := &shared.Node{
						OperationType: shared.POWER,
						Value:         0.0,
						Variable:      "",
						LNode: &shared.Node{
							OperationType: shared.VARIABLE,
							Value:         0.0,
							Variable:      key,
							LNode:         nil,
							RNode:         nil,
							Associative:   nil,
						},
						RNode: &shared.Node{
							OperationType: shared.NUMBER,
							Value:         fact,
							Variable:      "",
							LNode:         nil,
							RNode:         nil,
							Associative:   nil,
						},
						Associative: nil,
					}
					node.Associative = append(node.Associative, mult)
					changed = true
				}
			}
		}
		if changed && (nVarOp >= 1 || nNumOp >= 2) {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// x * (x + y + z) = x^2 + x*y + z*y
// (a+b) * (a+b) = a^2 + 2ab + b^2
func simplifyDefact(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MULTIPLY {
		// Exit if we are at the end of the tree.
		end := true
		for _, val := range node.Associative {
			if !isEndNode(val) {
				end = false
			}
		}
		if end {
			return nil, false, nil
		}

		res := &shared.Node{
			OperationType: shared.NUMBER,
			Value:         1.0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}

		for _, val := range node.Associative {
			res = multiplyNodes(res, val)
		}
		return res, true, nil
	}
	return nil, false, nil
}

// Finds the commen factor of all addends and puts it outside.
// abc + acd + ade -> a * (bc + cd + de)
// abc + a^2cd + ade -> a * (bc + acd + de)
func simplifyRefact(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.PLUS {
		// If there are less then 2 factors it doesn't matter if we refactor.
		if len(node.Associative) < 2 {
			return nil, false, nil
		}

		// resultshared.Node := clone(node)
		changed := false

		// Loop over all addends to search for common factor
		// We only search for one factor at a time so the methode terminates after it finds one.
		for _, val := range node.Associative {

			// Factor in front of the new addition term in the parenthesis.
			factors := &shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{},
			}

			// Addition Term inside the parenthesis.
			rest := &shared.Node{
				OperationType: shared.PLUS,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{},
			}

			// Rest of the addition that is not
			newAdditionNode := &shared.Node{
				OperationType: shared.PLUS,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{},
			}

			// If the addend is already a multiplication, search for each factor individually.
			// Else just search for the common factor.
			if val.OperationType != shared.MULTIPLY {

				// Search for the common factor.
				if ok, fact, newRest, newAddends := findCommonFactor(val, node.Associative); ok {

					// Prevent factoring with itself. i.e. x + y -> x * (1) + y
					if len(newRest) >= 2 {
						// Add the common factor into the multiplication term.
						factors.Associative = append(factors.Associative, fact)

						// Add the inside of the parenthesis term.
						rest.Associative = append(rest.Associative, newRest...)

						// Return the rest of the nodes that did not change by factoring.
						newAdditionNode.Associative = append(newAdditionNode.Associative, newAddends...)

						// Tell the program that a change in the term is required.
						changed = true
					}
				}
			} else {
				// Keep track of how the addends were already changed.
				// (abc, abd, abe) -> (bc, bd, be) -> (c, d, e)
				addends := node.Associative

				// Search for each factor individually.
				for _, factor := range val.Associative {

					// If we found a common factor.
					if ok, fact, newRest, newAddends := findCommonFactor(factor, addends); ok {
						if len(newRest) >= 2 {
							factors.Associative = append(factors.Associative, fact)
							rest.Associative = append(rest.Associative, newRest...)
							newAdditionNode.Associative = append(newAdditionNode.Associative, newAddends...)
							changed = true

							// Set the searchable set to the rest of the addition.
							// addends = newAdditionNode.Associative

							// Search for a maximum of one factor.
							break
						}
					}
				}
			}

			if len(factors.Associative) >= 1 && len(rest.Associative) >= 2 && changed {
				// changed = true
				// Devidor for all terms.
				divisor := &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         nil,
					RNode:         nil,
					Associative:   []*shared.Node{},
				}
				for _, fact := range factors.Associative {
					divisor.Associative = append(divisor.Associative, &shared.Node{
						OperationType: shared.POWER,
						Value:         0.0,
						Variable:      "",
						LNode:         fact,
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
				}

				// Add the divisor into every term in the parenthesis.
				for i, val := range rest.Associative {
					currentDivisor := clone(divisor)
					if val.OperationType == shared.MULTIPLY {
						currentDivisor.Associative = append(currentDivisor.Associative, val.Associative...)
					} else {
						currentDivisor.Associative = append(currentDivisor.Associative, val)
					}
					rest.Associative[i] = currentDivisor
				}
				factors.Associative = append(factors.Associative, rest)
				newAdditionNode.Associative = append(newAdditionNode.Associative, factors)
				return newAdditionNode, true, nil
			}
		}
		// if changed {
		//	 return resultshared.Node, true, nil
		// }
	}
	return nil, false, nil
}

// Check if a given factor appears in all factors. If it does, also return resulting rest.
func findCommonFactor(node *shared.Node, addends []*shared.Node) (bool, *shared.Node, []*shared.Node, []*shared.Node) {
	rest := []*shared.Node{}
	restAddends := []*shared.Node{}
	changed := false
	for _, val := range addends {
		if node.OperationType == shared.POWER {
			return findCommonFactor(node.LNode, addends)
		}

		if canFactor(node, val) {
			changed = true
			rest = append(rest, val)
		} else {
			restAddends = append(restAddends, val)
		}
	}
	if changed {
		return true, node, rest, restAddends
	} else {
		return false, nil, nil, nil
	}
}

// Returns the factor of a to make b. (factor, node) -> ok
// ab, abc -> true
// a, a^2b -> true
func canFactor(a, b *shared.Node) bool {
	switch b.OperationType {
	case shared.NUMBER:
		return a.OperationType == shared.NUMBER
	case shared.VARIABLE:
		return isEqual(a, b)
	case shared.MINUS:
		return canFactor(a, b.RNode)
	case shared.MULTIPLY:
		for _, val := range b.Associative {
			if canFactor(a, val) {
				return true
			}
		}
	case shared.POWER:
		return canFactor(a, b.LNode)
	default:
		return isEqual(a, b)
	}
	return false
}

// x^0 = 1.0
func simplifyPowZero(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.POWER {
		if isNumber(node.RNode) && isZero(node.RNode) {
			return &shared.Node{
				OperationType: shared.NUMBER,
				Value:         1.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, true, nil
		}
	}
	return nil, false, nil
}

// x * x = x^2
// Currently Unfunctional
func simplifyPowSelf(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MULTIPLY {
		if isEqual(node.RNode, node.LNode) {
			return &shared.Node{
				OperationType: shared.POWER,
				Value:         0.0,
				Variable:      "",
				LNode:         node.LNode,
				RNode: &shared.Node{
					OperationType: shared.NUMBER,
					Value:         2.0,
					Variable:      "",
					LNode:         nil,
					RNode:         nil,
					Associative:   nil,
				},
				Associative: nil,
			}, true, nil
		}
	}
	return nil, false, nil
}

// x^y + x^z = x^(y+z)
// Currently Unfunctional
func simplifyAddPow(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.MULTIPLY || node.OperationType == shared.DIVIDE {
		if node.LNode.OperationType == shared.POWER && node.RNode.OperationType == shared.POWER {
			if isEqual(node.LNode.LNode, node.RNode.LNode) {
				var op *shared.Node
				if node.OperationType == shared.MULTIPLY {
					op = &shared.Node{
						OperationType: shared.PLUS,
						Value:         0.0,
						Variable:      "",
						LNode:         node.LNode.RNode,
						RNode:         node.RNode.RNode,
						Associative:   nil,
					}
				} else {
					op = &shared.Node{
						OperationType: shared.MINUS,
						Value:         0.0,
						Variable:      "",
						LNode:         node.LNode.RNode,
						RNode:         node.RNode.RNode,
						Associative:   nil,
					}
				}
				return &shared.Node{
					OperationType: shared.POWER,
					Value:         0.0,
					Variable:      "",
					LNode:         node.LNode.LNode,
					RNode:         op,
					Associative:   nil,
				}, true, nil
			}
		}
	}
	return nil, false, nil
}

// (x^y)^z = x^(y*z)
func simplifyMultPow(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.POWER {
		if node.LNode.OperationType == shared.POWER {
			op := &shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   []*shared.Node{node.LNode.RNode, node.RNode}}
			return &shared.Node{
				OperationType: shared.POWER,
				Value:         0.0,
				Variable:      "",
				LNode:         node.LNode.LNode,
				RNode:         op,
				Associative:   nil,
			}, true, nil
		}
	}
	return nil, false, nil
}

// x * z + y * z = (x + y) * z
// Currently Unfunctional
func simplifyMultFact(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.PLUS || node.OperationType == shared.MINUS {
		if node.LNode.OperationType == shared.MULTIPLY && node.RNode.OperationType == shared.MULTIPLY {
			if isEqual(node.LNode.LNode, node.RNode.LNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.RNode),
					RNode:         clone(node.RNode.RNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.LNode,
					Associative:   nil,
				}, true, nil

			} else if isEqual(node.LNode.LNode, node.RNode.RNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.RNode),
					RNode:         clone(node.RNode.LNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.LNode,
					Associative:   nil,
				}, true, nil

			} else if isEqual(node.LNode.RNode, node.RNode.LNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.LNode),
					RNode:         clone(node.RNode.RNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.RNode,
					Associative:   nil,
				}, true, nil

			} else if isEqual(node.LNode.RNode, node.RNode.RNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.LNode),
					RNode:         clone(node.RNode.LNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.MULTIPLY,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.RNode,
					Associative:   nil,
				}, true, nil
			}
		}
	}
	return nil, false, nil
}

// x / z + y / z = (x + y) / z
// x / z + x / y = x / (z + y)
// Currently Unfunctional
func simplifyDivFact(node *shared.Node) (*shared.Node, bool, error) {
	if node.OperationType == shared.PLUS || node.OperationType == shared.MINUS {
		if node.LNode.OperationType == shared.DIVIDE && node.RNode.OperationType == shared.DIVIDE {
			if isEqual(node.LNode.LNode, node.RNode.LNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.RNode),
					RNode:         clone(node.RNode.RNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.DIVIDE,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.LNode,
					Associative:   nil,
				}, true, nil

			} else if isEqual(node.LNode.RNode, node.RNode.RNode) {
				op := shared.Node{
					OperationType: node.OperationType,
					Value:         0.0,
					Variable:      "",
					LNode:         clone(node.LNode.LNode),
					RNode:         clone(node.RNode.LNode),
					Associative:   nil,
				}
				return &shared.Node{
					OperationType: shared.DIVIDE,
					Value:         0.0,
					Variable:      "",
					LNode:         &op,
					RNode:         node.LNode.RNode,
					Associative:   nil,
				}, true, nil
			}
		}
	}
	return nil, false, nil
}

// eval everything that cannont produce an irational number
func simplifyConstantFold(node *shared.Node) (*shared.Node, bool, error) {
	switch node.OperationType {
	case shared.VARIABLE, shared.NUMBER:
		return nil, false, nil
	case shared.DIVIDE:
		if isNumber(node.LNode) && isNumber(node.RNode) {
			return &shared.Node{
				OperationType: shared.NUMBER,
				Value:         node.LNode.Value / node.RNode.Value,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, true, nil
		}
	case shared.MULTIPLY, shared.PLUS:
		res := 0.0
		for _, val := range node.Associative {
			if isNumber(val) {
				res += val.Value
			} else {
				return nil, false, nil
			}
		}

		if shared.Conf.Options["show_debug_process"] {
			cfmt.Printf("{{Debug:}}::cyan|bold All values are numbers.\n")
		}

		return &shared.Node{
			OperationType: shared.NUMBER,
			Value:         res,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}, true, nil
	case shared.MINUS:
		if isNumber(node.LNode) && isNumber(node.RNode) {
			return &shared.Node{
				OperationType: shared.NUMBER,
				Value:         node.LNode.Value - node.RNode.Value,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, true, nil
		}
	default:
	}
	return nil, false, nil
}
