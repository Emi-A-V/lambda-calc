package parser

import (
	"errors"
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

type parser struct {
	tokens       []shared.Token
	currentIndex int
}

func Parse(tokens []shared.Token) (shared.Node, error) {
	parser := parser{tokens, 0}
	node, err := parser.expr()
	if err != nil {
		return shared.Node{}, err
	}
	return node, nil
}

func (p *parser) expr() (shared.Node, error) {
	result, err := p.term()
	if err != nil {
		return shared.Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].TokenType == shared.PLUS {
			p.currentIndex += 1
			a, err := p.term()
			if err != nil {
				return shared.Node{}, err
			}
			c := result
			result = shared.Node{
				OperationType: shared.PLUS,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}
		} else if p.tokens[p.currentIndex].TokenType == shared.MINUS {
			p.currentIndex += 1
			a, err := p.term()
			if err != nil {
				return shared.Node{}, err
			}
			c := result
			result = shared.Node{
				OperationType: shared.MINUS,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}
		} else {
			break
		}
	}
	return result, nil
}

func (p *parser) term() (shared.Node, error) {
	result, err := p.factor()
	if err != nil {
		return shared.Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].TokenType == shared.MULTIPLY {
			p.currentIndex += 1
			a, err := p.factor()
			if err != nil {
				return shared.Node{}, err
			}
			c := result
			result = shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}
		} else if p.tokens[p.currentIndex].TokenType == shared.DIVIDE {
			p.currentIndex += 1
			a, err := p.factor()
			if err != nil {
				return shared.Node{}, err
			}
			c := result
			result = shared.Node{
				OperationType: shared.DIVIDE,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}

		} else if p.tokens[p.currentIndex].TokenType == shared.VARIABLE {
			c := result
			result = shared.Node{
				OperationType: shared.MULTIPLY,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode: &shared.Node{
					OperationType: shared.VARIABLE,
					Value:         0.0,
					Variable:      p.tokens[p.currentIndex].Variable,
					LNode:         nil,
					RNode:         nil,
					Associative:   nil,
				},
				Associative: nil,
			}
			p.currentIndex += 1
		} else {
			break
		}
	}
	return result, nil
}

func (p *parser) factor() (shared.Node, error) {
	result, err := p.num()
	if err != nil {
		return shared.Node{}, err
	}

	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].TokenType == shared.POWER {
			p.currentIndex += 1
			a, err := p.num()
			if err != nil {
				return shared.Node{}, err
			}
			c := result
			result = shared.Node{
				OperationType: shared.POWER,
				Value:         0.0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}
		} else {
			break
		}
	}
	return result, nil
}

func (p *parser) num() (shared.Node, error) {
	if p.currentIndex >= len(p.tokens) {
		cfmt.Println("{{Error:}}::red|bold unable to parse tokens, expecting another token.")
		return shared.Node{}, errors.New("missing token")
	}

	token := p.tokens[p.currentIndex]
	switch token.TokenType {
	case shared.NUMBER:
		a := shared.Node{
			OperationType: shared.NUMBER,
			Value:         p.tokens[p.currentIndex].Value,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}
		p.currentIndex += 1
		return a, nil
	case shared.VARIABLE:
		p.currentIndex += 1
		return shared.Node{
			OperationType: shared.VARIABLE,
			Value:         0.0,
			Variable:      p.tokens[p.currentIndex-1].Variable,
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}, nil
	case shared.LPARENTHESES:
		p.currentIndex += 1
		a, err := p.expr()
		if err != nil {
			return shared.Node{}, err
		} else if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].TokenType != shared.RPARENTHESES {
			cfmt.Println("{{Error:}}::red|bold unable to parse tokens, missing closing parentheses")
			return shared.Node{}, errors.New("unclosed parentheses")
		}
		p.currentIndex += 1
		return a, err
	case shared.SQRT:
		p.currentIndex += 1

		// Standard number for sqrt = 2
		a := shared.Node{
			OperationType: shared.NUMBER,
			Value:         2.0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}

		if p.currentIndex >= len(p.tokens) {
			cfmt.Println("{{Error:}}::red|bold unable to parse tokens, expecting another token")
			return shared.Node{}, errors.New("missing token")
		} else if p.tokens[p.currentIndex].TokenType == shared.POWER {
			p.currentIndex += 1
			var err error
			a, err = p.factor()
			if err != nil {
				return shared.Node{}, err
			}
			// p.currentIndex += 1
		}

		// Check for parentheses
		if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].TokenType != shared.LPARENTHESES {
			cfmt.Println("{{Error:}}::red|bold unable to parse tokens, missing opening parentheses")
			return shared.Node{}, errors.New("unopened parentheses")
		}
		p.currentIndex += 1

		// Get expression inside the root
		b, err := p.expr()
		if err != nil {
			return shared.Node{}, err
		}

		// Check for parentheses
		if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].TokenType != shared.RPARENTHESES {
			cfmt.Println("{{Error:}}::red|bold unable to parse tokens, missing closing parentheses")
			return shared.Node{}, errors.New("unclosed parentheses")
		}

		p.currentIndex += 1
		return shared.Node{
			OperationType: shared.SQRT,
			Value:         0.0,
			Variable:      "",
			LNode:         &a,
			RNode:         &b,
			Associative:   nil,
		}, nil
	case shared.PLUS:
		p.currentIndex += 1
		a, err := p.factor()
		if err != nil {
			return shared.Node{}, err
		}
		return shared.Node{
			OperationType: shared.PLUS,
			Value:         0.0,
			Variable:      "",
			LNode: &shared.Node{
				OperationType: shared.NUMBER,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			},
			RNode:       &a,
			Associative: nil,
		}, err
	case shared.MINUS:
		p.currentIndex += 1
		a, err := p.factor()
		if err != nil {
			return shared.Node{}, err
		}
		return shared.Node{
			OperationType: shared.MINUS,
			Value:         0.0,
			Variable:      "",
			LNode: &shared.Node{
				OperationType: shared.NUMBER,
				Value:         0.0,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			},
			RNode:       &a,
			Associative: nil,
		}, err
	default:
	}
	cfmt.Println("{{Error:}}::red|bold Unable to parse expression, unexpected token.")
	return shared.Node{}, errors.New("unexpected token")
}
