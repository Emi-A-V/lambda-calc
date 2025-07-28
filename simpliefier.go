package main

import (
	"errors"
	"slices"

	"github.com/i582/cfmt/cmd/cfmt"
)


type RewriteRule func(*Node) (*Node, bool, error)

// Constant index in rule set
const (
	UNWIND = 0
	REWIND = 1
	SOLVE = 2
)

var UnwindRules = []RewriteRule {
	// Basic elimination
	simplifyAddZero,		// 0
	simplifySubZero,		// 1
	simplifySingleAdd,	// 2
	simplifyMultZero,		// 3
	simplifyMultOne,		// 4
	simplifyDivOne,		 	// 5
	simplifyZeroDiv,		// 6
	simplifyDivSelf,		// 7
	
	// Power Rules
	// simplifyPowSelf,
	// simplifyAddPow,
	simplifyPowZero,		// 8
	simplifyMultPow,		// 9

	// Eval
	simplifyAddCollect,		// 10
	simplifyMultCollect,	// 11
	simplifyDefact,				// 12
	simplifyConstantFold,	// 13
}

var RewindRules = []RewriteRule {
	simplifyAddZero,		// 0
	simplifySubZero,		// 1
	simplifySingleAdd,	// 2
	simplifyMultZero,		// 3
	simplifyMultOne,		// 4
	simplifyDivOne,		 	// 5
	simplifyZeroDiv,		// 6
	simplifyDivSelf,		// 7

	simplifyRefact,
	simplifyAddCollect,		// 10
	simplifyMultCollect,	// 12
	simplifyConstantFold,	// 13
	simplifyPowZero,		// 8
	simplifyMultPow,		// 9
}

var SolveRules = []RewriteRule {
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
	UnwindRules,
	RewindRules,
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
		for i := 0; i < len(node.associative); i++ {
			val := node.associative[i]
			simp := &Node{}
			simp, err = simplify(val, mode)
			if err != nil {
				return nil, err
			}
			if simp.operationType == node.operationType {

				node.associative = removeFromNodeArray(node.associative, i)
				i --
				node.associative = append(node.associative, simp.associative...)
			} else {
				node.associative[i] = simp
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

	for i, rule := range RuleSets[mode] {
		if newNode, changed, err := rule(node); changed && err == nil {
			
			// Debug
			if config.Options["show_debug_process"] {
				cfmt.Printf("{{Notice:}}::blue|bold matched rule pattern %v, changed: ", i)
				printATree(node)
				cfmt.Printf(" to ")
				printATree(newNode)
				cfmt.Println("")
			} 
			
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
			} else if op == PLUS && n.operationType == MINUS {
				walk(n.lNode)
				result = append(result, &Node{MINUS, 0.0, "", &Node{NUMBER, 0.0, "", nil, nil, nil}, atr(n.rNode), nil})
			} else if op == MULTIPLY && n.operationType == DIVIDE {
				walk(n.lNode)
				result = append(result, &Node{POWER, 0.0, "", atr(n.rNode), &Node{NUMBER, -1.0, "", nil, nil, nil}, nil})
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
	} else if a.operationType == VARIABLE {
		return a.variable == b.variable
	} else if a.operationType == MULTIPLY || a.operationType == PLUS {
		return containSameNodes(a.associative, b.associative)
	}
	return true
}

// Similar to isEqual, but it returns true if a and b are factors of each other
// 1, 2 -> true, 2
// a, 2a -> true, 2
// a, a -> true, 1
// a, b -> false, 0
// abcd, abcd -> true, 1
// 2ab, ab -> true, 0.5
// a + b, 2a + 2b -> true, 2
func getMultiple(a, b *Node) (bool, float64) {
	if a.operationType != b.operationType {
		if a.operationType == VARIABLE && b.operationType == MULTIPLY {
			if ok, factor, variable := getNumFactor(b); ok {
				if variable.variable == a.variable {
					return true, factor.value
				}
			}
		} else if b.operationType == VARIABLE && a.operationType == MULTIPLY {
			if ok, factor, variable := getNumFactor(a); ok {
				if variable.variable == b.variable {
					return true, 1 / factor.value
				}
			}
		} else if a.operationType == MINUS {
			ok, x := getMultiple(a.rNode, b)
			return ok, x * -1 
		} else if b.operationType == MINUS {
			ok, x := getMultiple(a, b.rNode)
			return ok, x * -1
		}
		return false, 0.0
	
	// Check 
	} else if a.operationType == MULTIPLY {
		factor := 1.0
		// Map for checking if a Node already appeared in the other term.
		alreadySeenB := make(map[*Node]int)

		if config.Options["show_debug_process"] {
			cfmt.Printf("(simplifier - 237:4 - getMultiple) {{Debug:}}::cyan|bold Searching for factor between: ")
			printATree(a)
			cfmt.Printf(" and ")
			printATree(b)
			cfmt.Printf(".\n")
		}

		// Multiply the numbers in term b to the result factor and add all other factors to the alreadySeen map.
		// We shouldn't be able to see the same factor twice, because previously we simplified all duplicate factors to powers?
		for _, bVal := range b.associative {
			if bVal.operationType == NUMBER {
				factor = factor * bVal.value
			} else {
				alreadySeenB[bVal] = 0
			}
		}

		// For each value in the term a, change the result factor or search for the equal in the term b.
		for _, aVal := range a.associative {
			found := false


			// If current aVal is a Number divide the factor.
			if aVal.operationType == NUMBER {
				factor = factor / aVal.value
				continue
			}
			
			// Else search for the equal factor
			for _, bVal := range b.associative {
				// Skip if we see a number.
				if bVal.operationType == NUMBER {
					continue
				}

				// If we have not seen the value already and it is equal to aVal. 
				if alreadySeenB[bVal] < 1 {
					if isEqual(aVal, bVal) {
						// Add it to already seen so it is not checked again later.
						alreadySeenB[bVal]++
						found = true
						if config.Options["show_debug_process"] {
							cfmt.Printf("(simplifier - 289:6 - getMultiple) {{Debug:}}::cyan|bold Found factor: ")
							printATree(aVal)
							cfmt.Printf(" in ")
							printATree(b)
							cfmt.Printf("\n")
						}
					}
				}
			}
			// If we did not find a value in term b and it is not a number, we do not have a multple of the other term.
			if !found {
				if config.Options["show_debug_process"] {
					cfmt.Printf("(simplifier - 289:6 - getMultiple) {{Debug:}}::cyan|bold Did not find factor: ")
					printATree(aVal)
					cfmt.Printf(" in ")
						printATree(b)
					cfmt.Printf(".\n")
				}
				return false, 0
			}
		}
		
		// After we have searched for all values of 
		for _, bVal := range b.associative {
			if bVal.operationType == NUMBER {
				continue
			}
			// If we have not seen a factor of term b in a, return false.
			if alreadySeenB[bVal] < 1 {
				if config.Options["show_debug_process"] {
					cfmt.Printf("(simplifier - 287:6 - getMultiple) {{Debug:}}::cyan|bold Did not find factor: ")
					printATree(bVal)
					cfmt.Printf(" in ")
					printATree(a)
					cfmt.Printf(".\n")
				}
				return false, 0
			}
		}
		// Else return true and the factor.
		return true, factor
	} else if a.operationType == PLUS {
		factor := 1.0
		isFactorDefined := false
  	used := make(map[*Node]bool)

		for _, x := range b.associative {
			if x.operationType == NUMBER {
				factor = factor * x.value
			}
		}

		for _, x := range a.associative {
			if x.operationType == NUMBER {
				factor = factor / x.value
				continue
			}

			contains := false
			for _, y := range b.associative {

				if y.operationType == NUMBER {
					continue
				}

				if _, ok := used[y]; ok {
					continue
				}
			
				if ok, fact := getMultiple(x,y); ok {
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
					}else {
						return false, 0
					}
				}
			}
			if !contains {
				return false, 0.0
			}
		}
		return true, factor

	} else if a.operationType == NUMBER {
		return true, a.value / b.value
	}
	return false, 0.0
}

// Returns wether there is a factor of a variable, and if so than it also returns the factor and the variable.
// -> isAFactor, factor, variable
func getNumFactor(node *Node) (bool, *Node, *Node) {
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
	appended := []*Node{}

	for j,val := range a {
		if j == i {

		} else {
			appended = append(appended, val)
		}
	}
	return appended
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

// Checks if a node is either PLUS or MULTIPLY (Cascadable Operation)
func isEndNode(node *Node) bool {
	switch node.operationType {
	case PLUS, MULTIPLY:
		return false
	default:
		return true
	}
}

// Apply Multiplycation
// a * b = ab
// a * (a + b) = a * a + a * b
//
func multiplyNodes(x *Node, y *Node) *Node {
	
	if config.Options["show_debug_process"] {
		cfmt.Printf("(Simplifier - 275:6 - multiplyNodes) {{Debug:}}::cyan|bold Multiplying ")
 		printATree(x)
		cfmt.Printf(" and ")
		printATree(y)
		cfmt.Printf(" to ")
	}
	
	if x.operationType == NUMBER && x.value == 1 {
		if config.Options["show_debug_process"] {
 			printATree(y)
			cfmt.Printf("\n")
		}
		return y
	} else if y.operationType == NUMBER && y.value == 1 {
		if config.Options["show_debug_process"] {
 			printATree(x)
			cfmt.Printf("\n")
		}
		return x
	}


	res := &Node{PLUS, 0.0, "", nil, nil, []*Node{}}
	
	// Both are at the end of operation
	if isEndNode(x) && isEndNode(y) {
		res = &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{x,y}}

	// x is added into the multiply operation of y
	} else if y.operationType == MULTIPLY && isEndNode(x) {
		y.associative = append(y.associative, x)
		res = y
	
	// y is added into the multiply operation of x
	} else if x.operationType == MULTIPLY && isEndNode(y) {
		x.associative = append(x.associative, y)
		res = x
	
	// x is multiplied by every number in the y operation 
	} else if y.operationType == PLUS && isEndNode(x) {
		for _,val := range y.associative {
			res.associative = append(res.associative, &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{x, val}})
		}

	// x is multiplied by every number in the y operation 
	} else if x.operationType == PLUS && isEndNode(y) {
		for _,val := range x.associative {
			res.associative = append(res.associative, &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{y, val}})
		}
	
	}	else if x.operationType == PLUS && y.operationType == MULTIPLY {
		for _,val := range x.associative {
			a := clone(y)
			a.associative = append(a.associative, val)
			res.associative = append(res.associative, a)
		}
	}	else if x.operationType == MULTIPLY && y.operationType == PLUS {
		for _,val := range y.associative {
			a := clone(x)
			a.associative = append(a.associative, val)
			res.associative = append(res.associative, a)
		}
	
	// If both operations are not an end node, every value is multiplied.
	}else {
		for _, xVal := range x.associative {
			for _, yVal := range y.associative {
				res.associative = append(res.associative, &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{xVal,yVal}})
			}
		}
	}
	if config.Options["show_debug_process"] {
 		printATree(res)
		cfmt.Printf("\n")
	}
	return res
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
		for i := 0; i < len(node.associative); i++ {
			val := node.associative[i]
			if val.operationType == NUMBER && val.value == 1.0 {
				changed = true
				node.associative = removeFromNodeArray(node.associative, i)
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
				cfmt.Printf("{{Error:}}::red|bold Unable to simplify calculation, possible devision by zero.\n")
				return nil, false, errors.New("divide by 0")
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
				if varMap[val.variable] != 1 {
					nVarOp ++
				}
				i--
			case MINUS:
				switch val.rNode.operationType {
				case VARIABLE: 
					varMap[val.rNode.variable]--
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.variable] != -1 {
						nVarOp += 2
					}
					i--
				}
			case MULTIPLY:
				num := 1.0
				newVal := clone(val)
				
				// Knowing that if we found a previous multiple of this term, it should already be simplified.
				// Searching for next multiple of the current term.
				for y := i + 1; y < len(node.associative); y++ {
					// If we find a term that is a multiple..
					if ok, fact := getMultiple(val, node.associative[y]); ok {
						if config.Options["show_debug_process"] {
							cfmt.Printf("(Simplifier - 648:8 - simplifyAddCollect) {{Debug:}}::cyan|bold Found factor in addends: ")
							printATree(val)
							cfmt.Printf(" / ")
							printATree(node.associative[y])
							cfmt.Printf(" = ")
							cfmt.Printf("%v\n", fact)
						}
						num += fact
						node.associative = removeFromNodeArray(node.associative, y)
						y--
						nVarOp += 2
						changed = true
					}
				}
				if num != 1 {
					newVal.associative = append(newVal.associative, &Node{NUMBER, num, "", nil, nil, nil})
				}			
				node.associative[i] = newVal
			}
		}
		if result != 0.0 {
			node.associative = append(node.associative, &Node{NUMBER, result, "", nil, nil, nil})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				switch fact {
				case 0:
					changed = true
				case 1:
					node.associative = append(node.associative, &Node{VARIABLE, 0.0, key, nil, nil, nil})
					changed = true
				default:
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

// Collect all terms in `MULTIPLY` operations.
// a * a * b * 2 * 5 = 10 * b * a^2
func simplifyMultCollect(n *Node) (*Node, bool, error) {
	if n.operationType == MULTIPLY {
		node :=	clone(n)
		changed := false
		
		nNumOp := 0
		nVarOp := 0
		
		result := 1.0
		varMap := make(map[string]float64)

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
				if varMap[val.variable] != 1 {
					nVarOp++
				}
				i--
			case MINUS:
				if val.rNode.operationType == VARIABLE {
					varMap[val.rNode.variable]++
					result *= -1.0
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.rNode.variable] != 1 {
						nVarOp++
					}
					nNumOp += 2
					i--
				}
			case POWER:
				if val.lNode.operationType == VARIABLE && val.rNode.operationType == NUMBER {
					varMap[val.lNode.variable] += val.rNode.value
					node.associative = removeFromNodeArray(node.associative, i)
					if varMap[val.lNode.variable] != val.rNode.value {
						nVarOp ++
					}
					i--

				} else if val.lNode.operationType == NUMBER && val.rNode.operationType == NUMBER && val.rNode.value == -1 {
					result = result / val.lNode.value
					node.associative = removeFromNodeArray(node.associative, i)
					nNumOp++
					i--
				}
			}
		}
		if nNumOp >= 2 && nVarOp >= 2 && result == 1.0 {
			changed = true
		} else if result != 1 {
			node.associative = append(node.associative, &Node{NUMBER, result, "", nil, nil, nil})
			changed = true
		}
		if len(varMap) >= 1 {
			for key, fact := range varMap {
				if fact == 1 {
					node.associative = append(node.associative, &Node{VARIABLE, 0.0, key, nil, nil, nil})
					changed = true
				// If the Variables multiply together to x^0, replace them with 1
				} else if fact == 0 {
					node.associative = append(node.associative, &Node{NUMBER, 1.0, "", nil, nil, nil})
					changed = true
				} else {
					mult := &Node{POWER, 0.0, "", &Node{VARIABLE, 0.0, key, nil, nil, nil}, &Node{NUMBER, fact, "", nil, nil, nil}, nil}
					node.associative = append(node.associative, mult)	
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
func simplifyDefact(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		// Exit if we are at the end of the tree.
		end := true 
		for _, val := range node.associative {
			if !isEndNode(val) {
				end = false
			}
		}
		if end {
			return nil, false, nil
		}

		res := &Node{NUMBER, 1.0, "", nil, nil, nil}

		for _, val := range node.associative {
			res = multiplyNodes(res, val)
		}
		return res, true, nil
	}
	return nil, false, nil
}



// Finds the commen factor of all addends and puts it outside.
// abc + acd + ade -> a * (bc + cd + de)
// abc + a^2cd + ade -> a * (bc + acd + de)
func simplifyRefact(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS {
		if len(node.associative) <= 1 {
			return nil, false, nil
		}
		factors := &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{}}

		if config.Options["show_debug_process"] {
			cfmt.Printf("(Simplifier - 872:4 - simplifyRefact) {{Debug:}}::cyan|bold Searching for greatest common factor in: ")
			printATree(node)
			cfmt.Printf("\n")
		}

		val := node.associative[0]

		if val.operationType != MULTIPLY {
			if ok, fact := findCommonFactor(val, node.associative); ok {
				if config.Options["show_debug_process"] {
					cfmt.Printf("    Found factor: ")
					printATree(fact)
					cfmt.Printf("\n")
				}
				factors.associative = append(factors.associative, fact)
			}
		} else {
			for _, factor := range val.associative {
				if ok, fact := findCommonFactor(factor, node.associative); ok {
					if config.Options["show_debug_process"] {
						cfmt.Printf("    Found subfactor: ")
						printATree(fact)
						cfmt.Printf(" in addend: ")
						printATree(val)
						cfmt.Printf("\n")
					}
					factors.associative = append(factors.associative, fact)
				}
			}
		}
		if config.Options["show_debug_process"] {
			cfmt.Printf("    Result Factor: ")
			printATree(factors)
			cfmt.Printf("\n")
		}

		if len(factors.associative) >= 1 {
			newNode := &Node{PLUS, 0.0, "", nil, nil, []*Node{}}
			for _, val := range node.associative {
				newSubNode := &Node{MULTIPLY, 0.0, "", nil, nil, []*Node{}}
				for _, fact := range factors.associative {
					newSubNode.associative = append(newSubNode.associative, &Node{POWER, 0.0, "", fact, &Node{NUMBER, -1.0, "", nil, nil, nil}, nil})
				}

				if val.operationType == MULTIPLY {
					newSubNode.associative = append(newSubNode.associative, val.associative...)
				} else {
					newSubNode.associative = append(newSubNode.associative, val)
				}
				newNode.associative = append(newNode.associative, newSubNode)
			}
			factors.associative = append(factors.associative, newNode)
			printATree(factors)
			return factors, true, nil
		}
		
	}
	return nil, false, nil
}

// Check if a given factor appears in all factors. If it does also return resulting rest.
func findCommonFactor(node *Node, addends []*Node) (bool, *Node) {
	for _, val := range addends {
		if config.Options["show_debug_process"] {
			cfmt.Printf("    Searching factor: ")
			printATree(node)
			cfmt.Printf(" in addend: ")
			printATree(val)
			cfmt.Printf("\n")
		}
		
		if node.operationType == POWER {
			return findCommonFactor(node.lNode, addends)
		}

		if !canFactor(node, val) {
			if config.Options["show_debug_process"] {
				cfmt.Printf("    Breaking Search, did not find: ")
				printATree(node)
				cfmt.Printf(" in addend: ")
				printATree(val)
				cfmt.Printf("\n")
			}
			return false, nil
		}
	}
	return true, node 
}

// Returns the factor of a to make b. (factor, node) -> ok 
// ab, abc -> true
// a, a^2b -> true
func canFactor(a,b *Node) bool {
	switch b.operationType {
	case NUMBER:
		return a.operationType == NUMBER
	case VARIABLE:
		return isEqual(a,b)
	case MINUS:
		return canFactor(a, b.rNode)
	case MULTIPLY:
		for _, val := range b.associative {
			if canFactor(a,val) {
				return true
			}
		}
	case POWER:
		return canFactor(a, b.lNode)
	default:
		return isEqual(a, b)
	}
	return false
}



// x^0 = 1.0
func simplifyPowZero(node *Node) (*Node, bool, error) {
	if node.operationType == POWER {
		if isNumber(node.rNode) && isZero(node.rNode) {
			return &Node{NUMBER, 1.0, "", nil, nil, nil}, true, nil
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


// eval everything that cannont produce an irational number
func simplifyConstantFold(node *Node) (*Node, bool, error) {
	switch node.operationType {
	case VARIABLE, NUMBER:
		return nil, false, nil
	case DIVIDE:
		if isNumber(node.lNode) && isNumber(node.rNode) {
			return &Node{NUMBER, node.lNode.value / node.rNode.value, "", nil, nil, nil}, true, nil
		}
	case MULTIPLY, PLUS:
		res := 0.0
		for _, val := range node.associative {
			if isNumber(val) {
				res += val.value
			} else {
				return nil, false, nil
			}
		}

		if config.Options["show_debug_process"] {
			cfmt.Printf("{{Debug:}}::cyan|bold All values are numbers.\n")
		}
	
		return &Node{NUMBER, res, "", nil, nil, nil}, true, nil
	case MINUS:
		if isNumber(node.lNode) && isNumber(node.rNode) {
			return &Node{NUMBER, node.lNode.value - node.rNode.value, "", nil, nil, nil}, true, nil
		}
	default:
	}
	return nil, false, nil
}
