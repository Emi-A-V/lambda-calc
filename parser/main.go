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
	node, err := parser.assertion()
	if err != nil {
		return shared.Node{}, err
	}
	return node, nil
}

func (p *parser) assertion() (shared.Node, error) {
	result, err := p.expression()
	if err != nil {
		return shared.Node{}, err
	}
	if p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].TokenType == shared.EQUAL {
			p.currentIndex++
			a, err := p.expression()
			if err != nil {
				return shared.Node{}, err
			}
			c := result

			result = shared.Node{
				OperationType: shared.EQUAL,
				Value:         0,
				Variable:      "",
				LNode:         &c,
				RNode:         &a,
				Associative:   nil,
			}
		}
	}
	if p.currentIndex < len(p.tokens) {
		cfmt.Printf("{{Error:}}::bold|red Assertion statement is longer than expected.\n")
	}
	return result, nil
}

// Finds expressions of lowest associativity -> a + b | b - a
func (p *parser) expression() (shared.Node, error) {
	result, err := p.term()
	if err != nil {
		return shared.Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		switch p.tokens[p.currentIndex].TokenType {
		case shared.PLUS:
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
		case shared.MINUS:
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
		default:
			return result, nil
		}
	}
	return result, nil
}

// Finds expressions of middle associativity -> a * b | b / a
func (p *parser) term() (shared.Node, error) {
	result, err := p.factor()
	if err != nil {
		return shared.Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		switch p.tokens[p.currentIndex].TokenType {
		case shared.MULTIPLY:
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
		case shared.DIVIDE:
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

		case shared.VARIABLE:
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
		default:
			return result, nil
		}
	}
	return result, nil
}

// Finds expressions of highest associativity -> a ^ b
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

// Finds expressions of literal type or parenthesis.
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
		varName := p.tokens[p.currentIndex].Variable
		p.currentIndex += 1

		if p.currentIndex >= len(p.tokens) {
			// Cannot be a function. Return variable.
			return shared.Node{
				OperationType: shared.VARIABLE,
				Value:         0.0,
				Variable:      varName,
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, nil
		} else if p.tokens[p.currentIndex].TokenType != shared.LPARENTHESES {
			// Cannot be a function. Return variable.
			return shared.Node{
				OperationType: shared.VARIABLE,
				Value:         0.0,
				Variable:      varName,
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			}, nil
		}
		p.currentIndex += 1

		// Get expression inside the root
		prmt, err := p.parameter()
		if err != nil {
			return shared.Node{}, err
		}

		// Check for closing parentheses
		if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].TokenType != shared.RPARENTHESES {
			cfmt.Println("{{Error:}}::red|bold unable to parse tokens, missing closing parentheses")
			return shared.Node{}, errors.New("unclosed parentheses")
		}

		p.currentIndex += 1
		return shared.Node{
			OperationType: shared.FUNCTION,
			Value:         0.0,
			Variable:      varName,
			LNode:         nil,
			RNode:         nil,
			Associative:   prmt,
		}, nil
	case shared.LPARENTHESES:
		p.currentIndex += 1
		a, err := p.expression()
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
		b, err := p.expression()
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

func (p *parser) parameter() ([]*shared.Node, error) {
	parameters := []*shared.Node{}
	for {
		expr, err := p.expression()
		if err != nil {
			return []*shared.Node{}, err
		}
		parameters = append(parameters, &expr)
		if p.tokens[p.currentIndex].TokenType == shared.COMMA {
			p.currentIndex++
		} else {
			break
		}
	}
	return parameters, nil
}
