package lambdaengine

import (
	"errors"
	"slices"
)


type RewriteRule func(*Node) (*Node, bool, error)

// Constant index in rule set
const (
	NORMAL = 0
	SOLVE = 1
)

var NormalRules = []RewriteRule{
	// -- Basic elimination
	simplifyAddZero,
	simplifySubZero,
	simplifySingleAdd,
	simplifyMultZero,
	simplifyMultOne,
	simplifyDivOne,
	simplifyZeroDiv,
	simplifyDivSelf,
	
	// Power Rules
	// simplifyPowSelf,
	// simplifyAddPow,
	simplifyMultPow,

	// Factoring
	// simplifyDivFact,
	// simplifyMultFact,

	// Eval
	simplifyAddCollect,
	simplifyMultCollect,
	simplifyConstantFold,
}

var SolveRules = []RewriteRule{
	// Basic elimination
	simplifyAddZero,
	simplifySubZero,
	simplifyMultZero,
	simplifyMultOne,
	simplifyDivOne,
	simplifyZeroDiv,
	simplifyDivSelf,

	// Power Rules
	simplifyPowSelf,
	simplifyAddPow,
	simplifyMultPow,

	// Eval
	simplifyConstantFold,
}

var RuleSets = [][]RewriteRule{
	NormalRules,
	SolveRules,
}

// ----------------------------------- Main Methode ----------------------------------- 

func simplify(node *Node, mode int) (*Node, error) {
	if node == nil {
		return nil, nil
	}

	var err error
	switch node.operationType {
	case MULTIPLY, PLUS:
		for i, val := range node.associative {
			node.associative[i], err = simplify(val, mode)
			if err != nil {
				return nil, err
			}
		}
	case VARIABLE, NUMBER:
		return node, nil
	default:
		node.rNode, err = simplify(node.rNode, mode)
		if err != nil {
			return nil, err
		}

		node.lNode, err = simplify(node.lNode, mode)
		if err != nil {
			return nil, err
		}
	}

	for _, rule := range RuleSets[mode] {
		if newNode, changed, err := rule(node); changed && err == nil {
			return simplify(newNode, mode)
		} else if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// ----------------------------------- Associative Tree Rebuilder -----------------------------------
// The Associative Tree Rebuilder rebuilds the tree so that associative rules are followed, 
// i.e.: a + b + c = c + b + a
// After `atr` is used, all nodes of the type `MULTIPLY` and `PLUS`, use the `associative` variabe to 
// show all of its child nodes.

func atr(node *Node) *Node {
	switch node.operationType {
	case NUMBER, VARIABLE:
		return node
	case MINUS, PLUS, MULTIPLY, DIVIDE:
		var result []*Node
		var walk func(n *Node)
		var op int
		
		switch node.operationType {
		case MINUS:
			op = PLUS
		case DIVIDE:
			op = MULTIPLY
		default:
			op = node.operationType
		}

		walk = func(n *Node) {
			if n.operationType == op && n.operationType != MINUS && n.operationType != DIVIDE {
				walk(n.lNode)
				walk(n.rNode)
			} else if n.operationType == MINUS {
				walk(n.lNode)
				result = append(result, &Node{MINUS, 0.0, "", &Node{NUMBER, 0.0, "", nil, nil, nil}, atr(n.rNode), nil})
			} else if n.operationType == DIVIDE {
				walk(n.lNode)
				result = append(result, &Node{DIVIDE, 0.0, "", &Node{NUMBER, 1.0, "", nil, nil, nil}, atr(n.rNode), nil})
			} else {
				result = append(result, atr(n))
			}
		}
		walk(node)
		return &Node{op, 0.0, "", nil, nil, result}
	default:
		return &Node{node.operationType, 0.0, "", atr(node.lNode), atr(node.rNode), nil}
	}
}

// ----------------------------------- Helpers ----------------------------------- 

func isZero(n *Node) bool {
	return n.operationType == NUMBER && n.value == 0.0
}

func isNumber(n *Node) bool {
	return n.operationType == NUMBER
}

func isEqual(a, b *Node) bool {
	if a.operationType != b.operationType {
		return false
	} else if a.operationType == NUMBER {
		return a.value == b.value
	} else if a.operationType == DIVIDE {
		return isEqual(a.lNode, b.lNode) && isEqual(a.rNode, b.rNode)
	} else if a.operationType == MULTIPLY || a.operationType == PLUS {
		return containSameNodes(a.associative, b.associative)
	}
	return true
}

// Returns wether there is a factor of a variable, and if so than it also returns the factor and the variable.
// -> isAFactor, factor, variable
func getFactor(node *Node) (bool, *Node, *Node) {
	if node.operationType == MULTIPLY {
		if len(node.associative) == 2 {
			if node.associative[0].operationType == NUMBER && node.associative[1].operationType == VARIABLE {
				return true, node.associative[0], node.associative[1]
			} else if node.associative[0].operationType == VARIABLE && node.associative[1].operationType == NUMBER {
				return true, node.associative[1], node.associative[0]
			}
		}
	}
	return false, nil, nil
}

func clone(n *Node) *Node {
	if n == nil {
		return nil
	}

	copy := *n
	copy.lNode = clone(n.lNode)
	copy.rNode = clone(n.rNode)
	
	return &copy
}

func removeFromNodeArray(a []*Node, i int) []*Node {
	if len(a) - 1 == i {
		return a[:len(a) - 1]			
	}
	return append(a[:i], a[i+1:]...)
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
		for _, y := range a {

			// Skip if already used
			if _, ok := used[y]; ok {
				continue
			}
			
			// Check if equal
			if isEqual(x,y) {
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


// ----------------------------------- Rules -----------------------------------

// x + 0 = x
func simplifyAddZero(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS {
		changed := false
		for i, val := range node.associative {
			if isZero(val) {
				changed = true
				node.associative = removeFromNodeArray(node.associative, i)
			}
		}
		if changed {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// x - 0 = x
func simplifySubZero(node *Node) (*Node, bool, error) {
	if node.operationType == MINUS {
		if isZero(node.rNode) {
			return node.rNode, true, nil
		}
	}
	return nil, false, nil
}


// x * 0 = 0
func simplifyMultZero(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		if slices.ContainsFunc(node.associative, isZero) {
			return &Node{NUMBER, 0.0, "", nil, nil, nil}, true, nil
		}
	}
	return nil, false, nil
}


// x * 1 = x
func simplifyMultOne(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		changed := false
		for i, val := range node.associative {
			if val.operationType == NUMBER && val.value == 0.0 {
				changed = true
				node.associative = removeFromNodeArray(node.associative, i)
			}
		}
		if changed {
			return node, true, nil
		}
	}
	return nil, false, nil
}

// x / 1 = x
func simplifyDivOne(node *Node) (*Node, bool, error) {
	if node.operationType == DIVIDE {
		if isNumber(node.rNode) && node.rNode.value == 1 {
			return clone(node.lNode), true, nil
		}
	}
	return nil, false, nil
}

// 0 / x = 0 (x != 0)
// LATER: Maybe add assumptions
func simplifyZeroDiv(node *Node) (*Node, bool, error) {
	if node.operationType == DIVIDE {
		if isNumber(node.lNode) && node.lNode.value == 0 {
			if val, err := eval(node.rNode, true); err == nil && val != 0 {
				return &Node{NUMBER, 0.0, "", nil, nil, nil}, true, nil
			} else if val == 0 {
				return nil, false, errors.New("simplify division by zero")
			}
		}
	}
	return nil, false, nil
}

// x / x = 1
func simplifyDivSelf(node *Node) (*Node, bool, error) {
	if node.operationType == DIVIDE {
		if isEqual(node.rNode, node.lNode) {
			return &Node{ NUMBER, 1.0, "", nil, nil, nil} ,true, nil
		} 	
	}
	return nil, false, nil
}

// (+(a*b)) = (a*b)
// After collection of addition, clean up
func simplifySingleAdd(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MULTIPLY {
		if len(node.associative) == 1 {
			return node.associative[0], true, nil
		}
	}
	return nil, false, nil
}

// Collect all terms in `PLUS` operations.
// 2 + 4 + a + b + a = 6 + 2a + b
func simplifyAddCollect(n *Node) (*Node, bool, error) {
	if n.operationType == PLUS {
		node :=	clone(n)
		changed := false
		
		nNumOp := 0
		nVarOp := 0
		
		result := 0.0
		varMap := make(map[string]float64)

		for i := 0; i < len(node.associative); i++ {
			val := node.associative[i]
			switch val.operationType {
			case NUMBER:
				result += val.value
				node.associative = removeFromNodeArray(node.associative, i)
				nNumOp++
				i--
			case VARIABLE:
				varMap[val.variable]++
				node.associative = removeFromNodeArray(node.associative, i)
				if varMap[val.variable] >= 2 {
					nVarOp ++
				}
				i--
			case MINUS:
				if val.rNode.operationType == VARIABLE {
					varMap[val.rNode.variable]--
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.variable] >= 2 {
						nVarOp ++
					}
					i--
				}
			case MULTIPLY:
				if ok, factor, variable := getFactor(val); ok {
					varMap[variable.variable] += factor.value
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.variable] >= 2 {
						nVarOp ++
					}
					i--
				}
			}
		}
		if result != 0.0 {
			node.associative = append(node.associative, &Node{NUMBER, result, "", nil, nil, nil})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				if fact == 1 {
					node.associative = append(node.associative, &Node{VARIABLE, 0.0, key, nil, nil, nil})
					changed = true
				} else {
					mult := &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{ 
						{VARIABLE, 0.0, key, nil, nil, nil},
						{NUMBER, float64(fact), "", nil, nil, nil},
					}}
					node.associative = append(node.associative, mult)
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

// Collect all terms in `PLUS` operations.
// 2 + 4 + a + b + a = 6 + 2a + b
func simplifyMultCollect(n *Node) (*Node, bool, error) {
	if n.operationType == MULTIPLY {
		node :=	clone(n)
		changed := false
		
		nNumOp := 0
		nVarOp := 0
		
		result := 1.0
		varMap := make(map[string]int)

		for i := 0; i < len(node.associative); i++ {
			val := node.associative[i]
			switch val.operationType {
			case NUMBER:
				result *= val.value
				node.associative = removeFromNodeArray(node.associative, i)
				nNumOp++
				i--
			case VARIABLE:
				varMap[val.variable]++
				node.associative = removeFromNodeArray(node.associative, i)
				if varMap[val.variable] >= 2 {
					nVarOp ++
				}
				i--
			case MINUS:
				if val.rNode.operationType == VARIABLE {
					varMap[val.rNode.variable]++
					result *= -1.0
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.variable] >= 2 {
						nVarOp ++
					}
					i--
				}
			case DIVIDE:
				if val.rNode.operationType == VARIABLE {
					varMap[val.rNode.variable]--
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.variable] >= 2 {
						nVarOp ++
					}
					i--
				}
			}
		}
		if result != 1.0 {
			node.associative = append(node.associative, &Node{NUMBER, result, "", nil, nil, nil})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				if fact == 1 {
					node.associative = append(node.associative, &Node{VARIABLE, 0.0, key, nil, nil, nil})
					changed = true
				} else if fact == -1 { 
					// TODO: ^^^^^ Check if this is correct... Maybe it needs to go to the Power form?
					newNode := &Node{DIVIDE, 0.0, "", &Node{NUMBER, 1.0, "", nil, nil, nil}, &Node{VARIABLE, 0.0, key, nil, nil, nil}, nil}
					node.associative = append(node.associative, newNode)
					changed = true
				} else {
					mult := &Node{POWER, 0.0, "", &Node{VARIABLE, 0.0, key, nil, nil, nil}, &Node{NUMBER, float64(fact), "", nil, nil, nil}, nil}
					node.associative = append(node.associative, mult)
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


// x * x = x^2
// Currently Unfunctional
func simplifyPowSelf(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		if isEqual(node.rNode, node.lNode) {
			return &Node{POWER, 0.0, "", node.lNode, &Node{NUMBER, 2.0, "", nil, nil, nil}, nil}, true, nil
		}
	}
	return nil, false, nil
}

// x^y + x^z = x^(y+z)
// Currently Unfunctional
func simplifyAddPow(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY || node.operationType == DIVIDE {
		if node.lNode.operationType == POWER && node.rNode.operationType == POWER {
			if isEqual(node.lNode.lNode, node.rNode.lNode) {
				var op *Node
				if node.operationType == MULTIPLY {
					op = &Node{ PLUS, 0.0, "", node.lNode.rNode, node.rNode.rNode, nil}
				} else {
					op = &Node{ MINUS, 0.0, "", node.lNode.rNode, node.rNode.rNode, nil }
				}
				return &Node{POWER, 0.0, "", node.lNode.lNode, op, nil}, true, nil
			}
		}
	}
	return nil, false, nil
}

// (x^y)^z = x^(y*z)
func simplifyMultPow(node *Node) (*Node, bool, error) {
	if node.operationType == POWER {
		if node.lNode.operationType == POWER {
			op := &Node{ MULTIPLY, 0.0, "", nil, nil, []*Node{node.lNode.rNode, node.rNode}}
			return &Node{POWER, 0.0, "", node.lNode.lNode, op, nil}, true, nil
		}
	}
	return nil, false, nil
}

// x * z + y * z = (x + y) * z
// Currently Unfunctional
func simplifyMultFact(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MINUS {
		if node.lNode.operationType == MULTIPLY && node.rNode.operationType == MULTIPLY {
			if isEqual(node.lNode.lNode, node.rNode.lNode) {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.rNode), nil } 
				return &Node{MULTIPLY, 0.0, "", &op, node.lNode.lNode, nil }, true, nil

			} else if isEqual(node.lNode.lNode, node.rNode.rNode) {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.lNode), nil }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.lNode, nil }, true, nil
			
			} else if isEqual(node.lNode.rNode, node.rNode.lNode) {
				op := Node{node.operationType, 0.0, "", clone(node.lNode.lNode), clone(node.rNode.rNode), nil }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.rNode, nil }, true, nil

			} else if isEqual(node.lNode.rNode, node.rNode.rNode) {
				op := Node{ node.operationType,	0.0, "", clone(node.lNode.lNode), clone(node.rNode.lNode), nil }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.rNode, nil }, true, nil
			}
		}
	}
	return nil, false, nil
}

// x / z + y / z = (x + y) / z
// x / z + x / y = x / (z + y)
// Currently Unfunctional
func simplifyDivFact(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MINUS {
		if node.lNode.operationType == DIVIDE && node.rNode.operationType == DIVIDE {
			if isEqual(node.lNode.lNode, node.rNode.lNode) {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.rNode), nil }
				return &Node{ DIVIDE, 0.0, "", &op, node.lNode.lNode, nil }, true, nil
			
			} else if isEqual(node.lNode.rNode, node.rNode.rNode) {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.lNode), clone(node.rNode.lNode), nil }
				return &Node{ DIVIDE, 0.0, "", &op, node.lNode.rNode, nil}, true, nil	
			}
		}
	}
	return nil, false, nil
}


// (x + y) * z = x * z + y * z



// eval everything that cannont produce an irational number
func simplifyConstantFold(node *Node) (*Node, bool, error) {
	switch node.operationType {
	case VARIABLE, NUMBER, DIVIDE:
		return nil, false, nil
	case MULTIPLY, PLUS:
		for _, val := range node.associative {
			if !isNumber(val) {
				return nil, false, nil
			}
		}

		if val, err := eval(node, true); err == nil {	
			return &Node{NUMBER, val, "", nil, nil, nil}, true, nil
		} else {
			return nil, false, err
		}
	default:
		if isNumber(node.lNode) && isNumber(node.rNode) {
			if val, err := eval(node, true); err == nil {	
				return &Node{NUMBER, val, "", nil, nil, nil}, true, nil
			} else {
				return nil, false, err
			}
		}
	}
	return nil, false, nil
}
