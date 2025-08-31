package parser

import (
	"errors"
	"lambdacalc/shared"

	"github.com/i582/cfmt/cmd/cfmt"
)

// This is a rewrite of the main.go -> parse function.

type parser struct {
	currentToken *shared.Token
	tokens       []shared.Token
	index        int
}

const (
	ASSERTION = iota
	EUQALITY  = iota
)

// Only advance if there are tokens to read.
func (p *parser) advance() bool {
	p.index++
	if p.index >= len(p.tokens) {
		return false
	}
	p.currentToken = &p.tokens[p.index]
	return true
}

func (p *parser) hasNext() bool {
	return p.index < len(p.tokens)
}

func Parse(t []shared.Token) (*shared.Node, error) {
	parserObject := parser{
		currentToken: &t[0],
		tokens:       t,
		index:        0,
	}

	return parserObject.expression()
}

// Search Parse allows for searching differing top level patterns like
// assertions or equality checks, both being denoted with an equal sign
// in its structure.
func SearchParse(t []shared.Token, m int) (*shared.Node, error) {
	parserObject := parser{
		currentToken: &t[0],
		tokens:       t,
		index:        0,
	}

	return parserObject.topLevelStructures(m)
}

func (p *parser) topLevelStructures(mode int) (*shared.Node, error) {
	switch mode {
	case ASSERTION:
		// Get A part.
		a, err := p.expression()
		if err != nil {
			cfmt.Printf("{{Error:}}::red|bold Unable to parse assertion, fault assertion variable.\n")
			return nil, errors.New("faulty assertion front")
		}

		// Check if equal sign is there, then skip it.
		if p.currentToken.TokenType != shared.EQUAL {
			cfmt.Printf("{{Error:}}::red|bold Unable to parse assertion, mssing assertion symbol.\n")
			return nil, errors.New("expecting euqal symbol")
		} else if !p.advance() {
			cfmt.Printf("{{Error:}}::red|bold Unable to parse assertion, mssing assertion statement.\n")
			return nil, errors.New("missing assertion statement")
		}

		// Get B part.
		b, err := p.expression()
		if err != nil {
			cfmt.Printf("{{Error:}}::red|bold Unable to parse assertion, fault assertion equation.\n")
			return nil, errors.New("faulty assertion end")
		}

		if shared.Conf.Options["show_debug_process"] {
			cfmt.Printf("(parser 60:1 p.topLevelStructures) {{Debug:}}::cyan|bold Assigning %s the value %s.\n", shared.PrintATree(a), shared.PrintATree(b))
		}

		// Check if the equality statement is valid.

		return &shared.Node{
			OperationType: shared.EQUAL,
			Value:         0.0,
			Variable:      "",
			LNode:         a,
			RNode:         b,
			Associative:   nil,
		}, nil
	default:
		return p.expression()
	}
}

// Capture expressions with low associativity i.e.: x + y
// Iterorates as until the current item does not fit the pattern.
func (p *parser) expression() (*shared.Node, error) {
	addends := []*shared.Node{}
	operand := shared.PLUS

	for {
		// Capture higher level terms i.e.: x * y
		newFactor, err := p.term()
		if err != nil {
			return nil, err
		}

		if operand == shared.MINUS {
			newFactor = addMultiplication(newFactor, &shared.Node{
				OperationType: shared.NUMBER,
				Value:         -1,
				Variable:      "",
				LNode:         nil,
				RNode:         nil,
				Associative:   nil,
			})
		}
		addends = append(addends, newFactor)

		switch p.currentToken.TokenType {
		case shared.MINUS:
			operand = shared.MINUS
		case shared.PLUS:
			operand = shared.PLUS
		default:
			operand = 0
		}

		if operand == 0 {
			break
		} else if !p.advance() {
			break
		}
	}

	switch len(addends) {
	case 0:
		cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
		return nil, errors.New("missing token")
	case 1:
		return addends[0], nil
	default:
		return &shared.Node{
			OperationType: shared.PLUS,
			Value:         0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   addends,
		}, nil
	}
}

// Capture terms with mid associativity i.e.: x * y
func (p *parser) term() (*shared.Node, error) {
	factors := []*shared.Node{}
	operand := shared.MULTIPLY

	for {
		// Capture higher level terms i.e.: x  y
		newFactor, err := p.factor()
		if err != nil {
			return nil, err
		}

		if operand == shared.DIVIDE {
			newFactor = &shared.Node{
				OperationType: shared.POWER,
				Value:         0,
				Variable:      "",
				LNode:         newFactor,
				RNode: &shared.Node{
					OperationType: shared.NUMBER,
					Value:         -1.0,
					Variable:      "",
					LNode:         nil,
					RNode:         nil,
					Associative:   nil,
				},
				Associative: nil,
			}
		}
		factors = append(factors, newFactor)

		switch p.currentToken.TokenType {
		case shared.DIVIDE:
			operand = shared.DIVIDE
		case shared.MULTIPLY:
			operand = shared.MULTIPLY
		default:
			operand = 0
		}

		if operand == 0 {
			break
		} else if !p.advance() {
			break
		}
	}

	switch len(factors) {
	case 0:
		cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
		return nil, errors.New("missing token")
	case 1:
		return factors[0], nil
	default:
		return &shared.Node{
			OperationType: shared.MULTIPLY,
			Value:         0,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   factors,
		}, nil
	}
}

// Capture factors with high associativity i.e.: x ^ y
func (p *parser) factor() (*shared.Node, error) {
	var result *shared.Node
	var err error

	result, err = p.literal()
	if err != nil {
		return nil, err
	}

	for {
		if p.currentToken.TokenType == shared.POWER {
			if !p.advance() {
				cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
				return nil, errors.New("missing token")
			}

			exponent, err := p.literal()
			if err != nil {
				return nil, err
			}

			result = &shared.Node{
				OperationType: shared.POWER,
				Value:         0.0,
				Variable:      "",
				LNode:         result,
				RNode:         exponent,
				Associative:   nil,
			}
		} else {
			return result, nil
		}
	}
}

// Capture literals without associativity i.e.: x
func (p *parser) literal() (*shared.Node, error) {
	switch p.currentToken.TokenType {
	case shared.NUMBER:
		node := &shared.Node{
			OperationType: shared.NUMBER,
			Value:         p.currentToken.Value,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}
		if p.advance() {
		}
		return node, nil
	case shared.VARIABLE:
		varName := p.currentToken.Variable

		// If tokens following the variable match the pattern of a function.
		if p.advance() {
			if p.currentToken.TokenType == shared.LPARENTHESES {
				// Jump over the parenthesis.
				if !p.advance() {
					cfmt.Println("{{Error:}}::red|bold unable to parse tokens, expecting another token.")
					return nil, errors.New("missing token")
				}

				// Get all parameters of the function.
				parameters, err := p.parameter()
				if err != nil {
					return nil, err
				}

				if p.currentToken.TokenType != shared.RPARENTHESES {
					cfmt.Printf("{{Error:}}::bold|red unable to parse tokens, expecting closing parenthesis for function statement.")
					return nil, errors.New("unclosed function parenthesis")
				}

				// Check if the number of parameters matches a defined function.
				if val, ok := shared.Functions[varName]; ok {
					if len(val.Parameters) == len(parameters) {
					} else {
						cfmt.Printf("(parser 196:1 p.literal) {{Error:}}::bold|red Unable to parse tokens, incorrect amout of parameters.\n")
						return nil, errors.New("unmatched parameters")
					}
				}

				p.advance()
				return &shared.Node{
					OperationType: shared.FUNCTION,
					Value:         0.0,
					Variable:      varName,
					LNode:         nil,
					RNode:         nil,
					Associative:   parameters,
				}, nil
			}
		}

		// Else parse the token.
		return &shared.Node{
			OperationType: shared.VARIABLE,
			Value:         0.0,
			Variable:      varName,
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		}, nil
	case shared.LPARENTHESES:
		// Advancing over the parenthesis to analyse its contents.
		if !p.advance() {
			cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
			return nil, errors.New("missing token")
		}

		// Get expression inside the parenthesis.
		res, err := p.expression()
		if err != nil {
			return nil, err
		}

		// Advance and check for the closing parenthesis.
		if p.currentToken.TokenType == shared.RPARENTHESES {
			p.advance()
			return res, nil
		} else {
			cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting closing parenthesis.")
			return nil, errors.New("missing closing parenthesis")
		}
	case shared.PLUS:
		if !p.advance() {
			cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
			return nil, errors.New("missing token")
		}
		return p.literal()
	case shared.MINUS:
		if !p.advance() {
			cfmt.Println("{{Error:}}::red|bold Unable to parse tokens, expecting another token.")
			return nil, errors.New("missing token")
		}
		a, err := p.literal()
		if err != nil {
			return nil, err
		}
		node := addMultiplication(a, &shared.Node{
			OperationType: shared.NUMBER,
			Value:         -1,
			Variable:      "",
			LNode:         nil,
			RNode:         nil,
			Associative:   nil,
		})
		return node, nil
	default:
		cfmt.Printf("{{Error:}}::red|bold Unable to parse tokens, unexpected token: %v\n", p.currentToken)
		return nil, errors.New("unexpected token")
	}
}

func (p *parser) parameter() ([]*shared.Node, error) {
	parameters := []*shared.Node{}
	for {
		expr, err := p.expression()
		if err != nil {
			return []*shared.Node{}, err
		}
		parameters = append(parameters, expr)
		if p.tokens[p.index].TokenType == shared.COMMA {
			p.index++
		} else {
			break
		}
	}
	return parameters, nil
}

// Prevent from Multiplication being able to recurse.
// Add a node to childs multiplication.
func addMultiplication(parent *shared.Node, child *shared.Node) *shared.Node {
	if child.OperationType == shared.MULTIPLY && parent.OperationType != shared.MULTIPLY {
		child.Associative = append(child.Associative, parent)
		return child
	} else if child.OperationType != shared.MULTIPLY && parent.OperationType == shared.MULTIPLY {
		parent.Associative = append(parent.Associative, child)
		return parent
	} else if child.OperationType == shared.MULTIPLY && parent.OperationType == shared.MULTIPLY {
		child.Associative = append(child.Associative, parent.Associative...)
		return child
	}

	// Just make a new Multiplication and return it.
	node := &shared.Node{
		OperationType: shared.MULTIPLY,
		Value:         0,
		Variable:      "",
		LNode:         nil,
		RNode:         nil,
		Associative:   []*shared.Node{parent, child},
	}
	return node
}
