package interpreter

import (
	"errors"
	"lambdacalc/shared"
	"math"

	"github.com/i582/cfmt/cmd/cfmt"
)

func Evaluate(node *shared.Node, silent bool) (float64, error) {
	switch node.OperationType {
	case shared.NUMBER:
		return node.Value, nil

	case shared.VARIABLE:
		if val, ok := shared.Variables[node.Variable]; ok {
			a, err := Evaluate(&val, silent)
			if err != nil {
				return 0, err
			}
			return a, nil
		} else {
			cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, undefined variable '%s'.\n", node.Variable)
			return 0, errors.New("undefined variable")
		}
	case shared.PLUS:
		a := 0.0
		for _, val := range node.Associative {
			b, err := Evaluate(val, silent)
			if err != nil {
				return 0, err
			}
			a = a + b
		}
		return a, nil
	case shared.MINUS:
		a, err := Evaluate(node.LNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := Evaluate(node.RNode, silent)
		if err != nil {
			return 0, err
		}
		return a - b, nil
	case shared.MULTIPLY:
		a := 1.0
		for _, val := range node.Associative {
			b, err := Evaluate(val, silent)
			if err != nil {
				return 0, err
			}
			a = a * b
		}
		return a, nil
	case shared.DIVIDE:
		a, err := Evaluate(node.LNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := Evaluate(node.RNode, silent)
		if err != nil {
			return 0, err
		}
		if b == 0.0 {
			if !silent {
				cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, devision by zero.\n")
			}
			return 0, errors.New("divide by 0")
		}
		return a / b, nil
	case shared.POWER:
		a, err := Evaluate(node.LNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := Evaluate(node.RNode, silent)
		if err != nil {
			return 0, err
		}
		return math.Pow(a, b), nil
	case shared.SQRT:
		a, err := Evaluate(node.LNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := Evaluate(node.RNode, silent)
		if err != nil {
			return 0, err
		}
		if b <= 0 {
			if !silent {
				cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, result has no real solution.\n")
			}
			return 0, errors.New("negative sqrt")
		}
		return math.Pow(b, (1 / a)), nil
	default:
		if !silent {
			cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, unexpected symbole.\n")
		}
		return 0, errors.New("unexpected error")
	}
}
