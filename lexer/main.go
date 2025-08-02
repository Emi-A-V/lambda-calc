package lexer

import (
	"errors"
	"lambdacalc/shared"
	"strconv"
	"unicode"

	"github.com/i582/cfmt/cmd/cfmt"
)

func LexTokens(input string) ([]shared.Token, error) {
	i := 0
	var tokens []shared.Token
	for i < len(input) {
		switch rune(input[i]) {
		case []rune(shared.Conf.Symbols["plus"])[0]:
			token := shared.Token{
				TokenType: shared.PLUS,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["minus"])[0]:
			token := shared.Token{
				TokenType: shared.MINUS,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["multiply"])[0]:
			token := shared.Token{
				TokenType: shared.MULTIPLY,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["divide"])[0]:
			token := shared.Token{
				TokenType: shared.DIVIDE,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["power"])[0]:
			token := shared.Token{
				TokenType: shared.POWER,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["l_parentheses"])[0]:
			token := shared.Token{
				TokenType: shared.LPARENTHESES,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["r_parentheses"])[0]:
			token := shared.Token{
				TokenType: shared.RPARENTHESES,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["equal"])[0]:
			token := shared.Token{
				TokenType: shared.EQUAL,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		case []rune(shared.Conf.Symbols["parameter_split"])[0]:
			token := shared.Token{
				TokenType: shared.COMMA,
				Value:     0.0,
				Variable:  "",
			}
			tokens = append(tokens, token)
			i += 1
		default:
			// Decode Numbers
			if unicode.IsNumber(rune(input[i])) || rune(input[i]) == []rune(shared.Conf.Symbols["decimal_split"])[0] {
				dot := false
				str := ""
				for i < len(input) && (unicode.IsNumber(rune(input[i])) || rune(input[i]) == []rune(shared.Conf.Symbols["decimal_split"])[0]) {
					if dot && rune(input[i]) == []rune(shared.Conf.Symbols["decimal_split"])[0] {
						cfmt.Printf("{{Error:}}::red|bold Unable to parse number at %v, reading multiple decimal splits.\n", i)
						return nil, errors.New("multiple decimal splits")
					} else if rune(input[i]) == []rune(shared.Conf.Symbols["decimal_split"])[0] {
						dot = true
						str += "."
					} else {
						str += string(input[i])
					}
					i += 1
				}
				num, err := strconv.ParseFloat(str, 64)
				if err != nil {
					cfmt.Printf("{{Error:}}::red|bold Unable to parse number, character-conversion faild.\n")
					return nil, errors.New("number parsing")
				}
				tokens = append(tokens, shared.Token{
					TokenType: shared.NUMBER,
					Value:     num,
					Variable:  "",
				})
			} else if unicode.IsSpace(rune(input[i])) {
				// Skip empty space
				i += 1
			} else if unicode.IsLetter(rune(input[i])) {
				// Search for constants or variables
				str := ""

				for i < len(input) && unicode.IsLetter(rune(input[i])) {
					str += string(input[i])
					i += 1
				}

				// Check Constants
				if val, ok := shared.Conf.Constants[str]; ok {
					tokens = append(tokens, shared.Token{
						TokenType: shared.NUMBER,
						Value:     val,
						Variable:  "",
					})
				} else if str == "sqrt" {
					tokens = append(tokens, shared.Token{
						TokenType: shared.SQRT,
						Value:     0.0,
						Variable:  "",
					})
				} else {
					// Its neither a constant or a sqrt than break the string into single variables.
					j := 0
					for j < len(str) {
						// Debug
						if shared.Conf.Options["show_debug_process"] {
							if val, ok := shared.Variables[string(str[j])]; ok {
								cfmt.Printf("{{Notice:}}::blue|bold found defined variable %s with value: ", string(str[j]))
								cfmt.Printf("%s", shared.PrintATree(&val))
								cfmt.Printf(".\n")
							} else if val, ok := shared.Functions[string(str[j])]; ok {
								cfmt.Printf("{{Notice:}}::blue|bold found defined variable %s with value: ", string(str[j]))
								cfmt.Printf("%s", shared.PrintATree(val.Equation))
								cfmt.Printf(".\n")
							} else {
								cfmt.Printf("{{Notice:}}::blue|bold found undefined variable %s.\n", string(str[j]))
							}
						}
						tokens = append(tokens, shared.Token{
							TokenType: shared.VARIABLE,
							Value:     0.0,
							Variable:  string(str[j]),
						})
						j += 1
					}
				}
			} else {
				cfmt.Printf("{{Error:}}::red|bold Unable to parse symbol at %s, unregocnized character.\n", string(input[i]))
				return nil, errors.New("number parsing")
			}
		}
	}
	return tokens, nil
}
