package lambdaengine

import (
	"errors"
)


func parse(tokens []Token) (Node, error) {
	parser := Parser{tokens, 0}
	node, err := parser.expr()
	if err != nil {
		return Node{}, err
	}
	return node, nil
}

type Parser struct {
	tokens       []Token
	currentIndex int
}

func (p *Parser) expr() (Node, error) {
	result, err := p.term()
	if err != nil {
		return Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].tokenType == PLUS {
			p.currentIndex += 1
			a, err := p.term()
			if err != nil {
				return Node{}, err
			}
			c := result
			result = Node{PLUS, 0.0, "", &c, &a, nil}
		} else if p.tokens[p.currentIndex].tokenType == MINUS {
			p.currentIndex += 1
			a, err := p.term()
			if err != nil {
				return Node{}, err
			}
			c := result
			result = Node{MINUS, 0.0, "", &c, &a, nil}
		} else {
			break
		}
	}
	return result, nil
}

func (p *Parser) term() (Node, error) {
	result, err := p.factor()
	if err != nil {
		return Node{}, err
	}
	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].tokenType == MULTIPLY {
			p.currentIndex += 1
			a, err := p.factor()
			if err != nil {
				return Node{}, err
			}
			c := result
			result = Node{MULTIPLY, 0.0, "", &c, &a, nil}
		} else if p.tokens[p.currentIndex].tokenType == DIVIDE {
			p.currentIndex += 1
			a, err := p.factor()
			if err != nil {
				return Node{}, err
			}
			c := result
			result = Node{DIVIDE, 0.0, "", &c, &a, nil}

		} else if p.tokens[p.currentIndex].tokenType == VARIABLE {
			c := result
			result = Node{MULTIPLY, 0.0, "", &c, &Node{VARIABLE, 0.0, p.tokens[p.currentIndex].variable, nil, nil, nil}, nil}
			p.currentIndex += 1
		} else {
			break
		}
	}
	return result, nil
}

func (p *Parser) factor() (Node, error) {
	result, err := p.num()
	if err != nil {
		return Node{}, err
	}

	for p.currentIndex < len(p.tokens) {
		if p.tokens[p.currentIndex].tokenType == POWER {
			p.currentIndex += 1
			a, err := p.num()
			if err != nil {
				return Node{}, err
			}
			c := result
			result = Node{POWER, 0.0, "", &c, &a, nil}
		} else {
			break
		}
	}
	return result, nil
}

func (p *Parser) num() (Node, error) {
	if p.currentIndex >= len(p.tokens) {
		return Node{}, errors.New("missing token")
	}

	token := p.tokens[p.currentIndex]
	switch token.tokenType {
	case NUMBER:
		a := Node{NUMBER, p.tokens[p.currentIndex].value, "", nil, nil, nil}
		p.currentIndex += 1
		return a, nil
	case VARIABLE:
		p.currentIndex += 1
	 	return Node{VARIABLE, 0.0, p.tokens[p.currentIndex - 1].variable, nil, nil, nil}, nil
	case LPARENTHESES:
		p.currentIndex += 1
		a, err := p.expr()
		if err != nil {
			return Node{}, err
		} else if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].tokenType != RPARENTHESES {
			return Node{}, errors.New("unclosed parentheses")
		}
		p.currentIndex += 1
		return a, err
	case SQRT:
		p.currentIndex += 1

		// Standard number for sqrt = 2
		a := Node{NUMBER, 2.0, "", nil, nil, nil}

		if p.currentIndex >= len(p.tokens) {
			return Node{}, errors.New("missing token")
		} else if p.tokens[p.currentIndex].tokenType == POWER {
			p.currentIndex += 1
			var err error
			a, err = p.factor()
			if err != nil {
				return Node{}, err
			}
			// p.currentIndex += 1
		}

		// Check for parentheses
		if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].tokenType != LPARENTHESES {
			return Node{}, errors.New("unopened parentheses")
		}
		p.currentIndex += 1

		// Get expression inside the root
		b, err := p.expr()
		if err != nil {
			return Node{}, err
		}

		// Check for parentheses
		if p.currentIndex >= len(p.tokens) || p.tokens[p.currentIndex].tokenType != RPARENTHESES {
			return Node{}, errors.New("unclosed parentheses")
		}

		p.currentIndex += 1
		return Node{SQRT, 0.0, "", &a, &b, nil}, nil
	case PLUS:
		p.currentIndex += 1
		a, err := p.factor()
		if err != nil {
			return Node{}, err
		}
		return Node{PLUS, 0.0, "", &Node{NUMBER, 0.0, "", nil, nil, nil}, &a, nil}, err
	case MINUS:
		p.currentIndex += 1
		a, err := p.factor()
		if err != nil {
			return Node{}, err
		}
		return Node{MINUS, 0.0, "", &Node{NUMBER, 0.0, "", nil, nil, nil}, &a, nil}, err
	default:
	}
	return Node{}, errors.New("unexpected token")
}


