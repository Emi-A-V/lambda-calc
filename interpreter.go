package main

import (
	"github.com/i582/cfmt/cmd/cfmt"
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
			cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, undefined variable '%s'.\n", node.variable)
			return 0, errors.New("undefined variable")
		}
	case PLUS:
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		return a + b, nil
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
		a, err := eval(node.lNode, silent)
		if err != nil {
			return 0, err
		}
		b, err := eval(node.rNode, silent)
		if err != nil {
			return 0, err
		}
		return a * b, nil
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
			if !silent {
				cfmt.Printf("{{Error:}}::red|bold Unable to calculate output, devision by zero.\n")
			}
			return 0, errors.New("divide by 0")
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
