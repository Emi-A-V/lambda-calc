package simplifier

import (
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

// The simplify function reads a node and traverses along its branches to find rules of a specified ruleset to apply.
func Simplify(node *shared.Node, mode int) (*shared.Node, error) {
	if node == nil {
		return nil, nil
	}

	var err error
	switch node.OperationType {
	case shared.MULTIPLY, shared.PLUS, shared.FUNCTION:
		for i := 0; i < len(node.Associative); i++ {
			val := node.Associative[i]
			simp := &shared.Node{}
			simp, err = Simplify(val, mode)
			if err != nil {
				return nil, err
			}
			if simp.OperationType == node.OperationType {

				node.Associative = removeFromNodeArray(node.Associative, i)
				i--
				node.Associative = append(node.Associative, simp.Associative...)
			} else {
				node.Associative[i] = simp
			}
		}
	case shared.VARIABLE, shared.NUMBER:
		return node, nil
	default:
		node.RNode, err = Simplify(node.RNode, mode)
		if err != nil {
			return nil, err
		}

		node.LNode, err = Simplify(node.LNode, mode)
		if err != nil {
			return nil, err
		}
	}

	for i, rule := range RuleSets[mode] {
		if newNode, changed, err := rule(node); changed && err == nil {

			// Debug
			if shared.Conf.Options["show_debug_process"] {
				cfmt.Printf("{{Notice:}}::blue|bold matched rule pattern %v, changed: ", i)
				cfmt.Printf("%s", shared.PrintATree(node))
				cfmt.Printf(" to ")
				cfmt.Printf("%s", shared.PrintATree(newNode))
				cfmt.Println("")
			}

			return Simplify(newNode, mode)
			// return newNode, nil
		} else if err != nil {
			return nil, err
		}
	}

	return node, nil
}

// Rule Type.
type RewriteRule func(*shared.Node) (*shared.Node, bool, error)

// Constant index in rule set
const (
	UNWIND = 0
	REWIND = 1
	SOLVE  = 2
)

// Unwind rule try to bring the equation to point were all
var UnwindRules = []RewriteRule{
	// Basic elimination
	simplifyAddZero,   // 0
	simplifySubZero,   // 1
	simplifySingleAdd, // 2
	simplifyMultZero,  // 3
	simplifyMultOne,   // 4
	simplifyDivOne,    // 5
	simplifyZeroDiv,   // 6
	simplifyDivSelf,   // 7

	// Power Rules
	// simplifyPowSelf,
	// simplifyAddPow,
	simplifyPowZero, // 8
	simplifyMultPow, // 9

	// Eval
	simplifyAddCollect,   // 10
	simplifyMultCollect,  // 11
	simplifyDefact,       // 12
	simplifyConstantFold, // 13
}

var RewindRules = []RewriteRule{
	simplifyAddZero,   // 0
	simplifySubZero,   // 1
	simplifySingleAdd, // 2
	simplifyMultZero,  // 3
	simplifyMultOne,   // 4
	simplifyDivOne,    // 5
	simplifyZeroDiv,   // 6
	simplifyDivSelf,   // 7

	simplifyAddCollect,   // 10
	simplifyMultCollect,  // 12
	simplifyConstantFold, // 13
	simplifyPowZero,      // 8
	simplifyMultPow,      // 9
	simplifyRefact,
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
	UnwindRules,
	RewindRules,
	SolveRules,
}
