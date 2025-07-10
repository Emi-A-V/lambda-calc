package main

import (
	"errors"
	"sort"

	"github.com/i582/cfmt/cmd/cfmt"
	"golang.org/x/text/cases"
)


type RewriteRule func(*Node) (*Node, bool, error)

// Constant index in rule set
const (
	NORMAL = 0
	SOLVE = 1
)

var NormalRules = []RewriteRule{
	// Basic elimination
	simplifyAddZero,
	simplifyMultZero,
	simplifyMultOne,
	simplifyDivOne,
	simplifyZeroDiv,
	simplifyDivSelf,
	
	// Power Rules
	simplifyPowSelf,
	simplifyAddPow,
	simplifyMultPow,

	// Factoring
	simplifyDivFact,
	simplifyMultFact,

	// Eval
	simplifyConstantFold,
}

var SolveRules = []RewriteRule{
	// Basic elimination
	simplifyAddZero,
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
	node.rNode, err = simplify(node.rNode, mode)
	if err != nil {
		return nil, err
	}

	node.lNode, err = simplify(node.lNode, mode)
	if err != nil {
		return nil, err
	}


	for _, rule := range RuleSets[mode] {
		if newNode, changed, err := rule(node); changed && err == nil {
			
			// Debug
			if config.Options["show_debug_process"] {
				cfmt.Printf("{{Notice:}}::blue|bold matched rule pattern, changed: ")
				printTree(node)
				cfmt.Printf(" to ")
				printTree(newNode)
				cfmt.Println("")
			} 
			
			return simplify(newNode, mode)
		} else if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// ----------------------------------- Helpers ----------------------------------- 

func isZero(n *Node) bool {
	return n.operationType == NUMBER && n.value == 0
}

func isNumber(n *Node) bool {
	return n.operationType == NUMBER
}

func isEqual(a, b *Node) bool {

}

func normalizeCommunicative(node *Node) *Node {
	if node.operationType == MULTIPLY || node.operationType == PLUS {
		terms := collectTerms(node, node.operationType)		
		sort.Slice(terms, func(i, j int) bool {
			return termString
		})
	}
	return node
}

func collectTerms(node *Node, op int) []*Node {
	
}

func isStructEqual(a, b *Node) bool {
	if a.operationType != b.operationType {
		return false
	}

	switch a.operationType {
	case NUMBER:
		return a.value == b.value
	case VARIABLE:
		return a.variable == a.variable
	default:
		return a.operationType == b.operationType &&
		isStructEqual(a.lNode, b.lNode) &&
			isStructEqual(a.rNode, b.rNode)
	}
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


// ----------------------------------- Rules -----------------------------------

// x + 0 = x
// x - 0 = x
func simplifyAddZero(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MINUS {
		if isZero(node.rNode) {
			return clone(node.lNode), true, nil
		} else if isZero(node.lNode) {
			return clone(node.rNode), true, nil
		}
	}
	return nil, false, nil
}


// x * 0 = 0
func simplifyMultZero(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		if isZero(node.rNode) || isZero(node.lNode) {
			return &Node{NUMBER, 0.0, "", nil, nil}, true, nil
		}
	}
	return nil, false, nil
}


// x * 1 = x
func simplifyMultOne(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		if isNumber(node.rNode) && node.rNode.value == 1 {
			return clone(node.lNode), true, nil
		} else  if isNumber(node.lNode) && node.lNode.value == 1 {
			return clone(node.rNode), true, nil
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
func simplifyZeroDiv(node *Node) (*Node, bool, error) {
	if node.operationType == DIVIDE {
		if isNumber(node.lNode) && node.lNode.value == 0 {
			if val, err := eval(node.rNode, true); err == nil && val != 0 {
				return &Node{NUMBER, 0.0, "", nil, nil}, true, nil
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
		if *node.rNode == *node.lNode {
			return &Node{ NUMBER, 1.0, "", nil, nil } ,true, nil
		} 	
	}
	return nil, false, nil
}

// x * x = x^2
func simplifyPowSelf(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY {
		if *node.rNode == *node.lNode {
			return &Node{POWER, 0.0, "", node.lNode, &Node{NUMBER, 2.0, "", nil, nil}}, true, nil
		}
	}
	return nil, false, nil
}

// x^y + x^z = x^(y+z)
func simplifyAddPow(node *Node) (*Node, bool, error) {
	if node.operationType == MULTIPLY || node.operationType == DIVIDE {
		if node.lNode.operationType == POWER && node.rNode.operationType == POWER {
			if *node.lNode.lNode == *node.rNode.lNode {
				var op *Node
				if node.operationType == MULTIPLY {
					op = &Node{ PLUS, 0.0, "", node.lNode.rNode, node.rNode.rNode }
				} else {
					op = &Node{ MINUS, 0.0, "", node.lNode.rNode, node.rNode.rNode }
				}
				return &Node{POWER, 0.0, "", node.lNode.lNode, op}, true, nil
			}
		}
	}
	return nil, false, nil
}

// (x^y)^z = x^(y*z)
func simplifyMultPow(node *Node) (*Node, bool, error) {
	if node.operationType == POWER {
		if node.lNode.operationType == POWER {
			op := &Node{ MULTIPLY, 0.0, "", node.lNode.rNode, node.rNode }
			return &Node{POWER, 0.0, "", node.lNode.lNode, op}, true, nil
		}
	}
	return nil, false, nil
}

// x * z + y * z = (x + y) * z
func simplifyMultFact(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MINUS {
		if node.lNode.operationType == MULTIPLY && node.rNode.operationType == MULTIPLY {
			if *node.lNode.lNode == *node.rNode.lNode {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.rNode) }
				return &Node{MULTIPLY, 0.0, "", &op, node.lNode.lNode}, true, nil

			} else if *node.lNode.lNode == *node.rNode.rNode {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.lNode) }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.lNode }, true, nil
			
			} else if *node.lNode.rNode == *node.rNode.lNode {
				op := Node{node.operationType, 0.0, "", clone(node.lNode.lNode), clone(node.rNode.rNode) }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.rNode }, true, nil

			} else if *node.lNode.rNode == *node.rNode.rNode {
				op := Node{ node.operationType,	0.0, "", clone(node.lNode.lNode), clone(node.rNode.lNode) }
				return &Node{ MULTIPLY, 0.0, "", &op, node.lNode.rNode }, true, nil
			}
		}
	}
	return nil, false, nil
}

// x / z + y / z = (x + y) / z
// x / z + x / y = x / (z + y)
func simplifyDivFact(node *Node) (*Node, bool, error) {
	if node.operationType == PLUS || node.operationType == MINUS {
		if node.lNode.operationType == DIVIDE && node.rNode.operationType == DIVIDE {
			if *node.lNode.lNode == *node.rNode.lNode {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.rNode), clone(node.rNode.rNode) }
				return &Node{ DIVIDE, 0.0, "", &op, node.lNode.lNode }, true, nil
			
			} else if *node.lNode.rNode == *node.rNode.rNode {
				op := Node{ node.operationType, 0.0, "", clone(node.lNode.lNode), clone(node.rNode.lNode) }
				return &Node{ DIVIDE, 0.0, "", &op, node.lNode.rNode }, true, nil	
			}
		}
	}
	return nil, false, nil
}


// (x + y) * z = x * z + y * z



// eval everything that cannont produce an irational number
func simplifyConstantFold(node *Node) (*Node, bool, error) {
	if node.operationType != VARIABLE && 
		node.operationType != NUMBER &&
		node.operationType != DIVIDE {
		if isNumber(node.lNode) && isNumber(node.rNode) {
			if val, err := eval(node, true); err == nil {	
				return &Node{NUMBER, val, "", nil, nil}, true, nil
			} else {
				return nil, false, err
			}
		}
	}
	return nil, false, nil
}
