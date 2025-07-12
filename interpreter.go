package lambdacalc

import (
	"math"
	"errors"
)


func eval(node *Node, silent bool) (float64, error) {
	switch node.operationType {
	case NUMBER:
		return node.value, nil

	case VARIABLE:
		if val, ok := variables[node.variable]; ok {
			a, err := eval(&val, silent)
			if err != nil {
				return 0, err
			}
			return a, nil
		} else {
			return 0, errors.New("undefined variable")
		}
	case PLUS:
		a := 0.0
		for _, val := range node.associative {
			b, err := eval(val, silent)
			if err != nil {
				return 0, err
			}
			a = a + b
		}
		return a, nil
	case MINUS:
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		return a - b, nil
	case MULTIPLY:
		a := 1.0
		for _, val := range node.associative {
			b, err := eval(val, silent)
			if err != nil {
				return 0, err
			}
			a = a * b
		}
		return a, nil
	case DIVIDE:
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		if b == 0.0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	case POWER:
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		return math.Pow(a, b), nil
	case SQRT:
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		if b <= 0 {
			return 0, errors.New("negative sqrt")
		}
		return math.Pow(b, (1 / a)), nil
	default:
		return 0, errors.New("unexpected error")
	}
}
