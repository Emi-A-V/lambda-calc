package main

import (
	"github.com/i582/cfmt/cmd/cfmt"
	"unicode"
	"errors"
	"strconv"
)

func lexer(input string) ([]Token, error) {
	i := 0
	var tokens []Token
	for i < len(input) {
		switch rune(input[i]) {
		case []rune(config.Symbols["plus"])[0]:
			token := Token{PLUS, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["minus"])[0]:
			token := Token{MINUS, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["multiply"])[0]:
			token := Token{MULTIPLY, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["divide"])[0]:
			token := Token{DIVIDE, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["power"])[0]:
			token := Token{POWER, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["l_parentheses"])[0]:
			token := Token{LPARENTHESES, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["r_parentheses"])[0]:
			token := Token{RPARENTHESES, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		case []rune(config.Symbols["equal"])[0]:
			token := Token{EQUAL, 0.0, ""}
			tokens = append(tokens, token)
			i += 1
		default:
			// Decode Numbers
			if unicode.IsNumber(rune(input[i])) || rune(input[i]) == []rune(config.Symbols["decimal_split"])[0] {
				dot := false
				str := ""
				for i < len(input) && (unicode.IsNumber(rune(input[i])) || rune(input[i]) == []rune(config.Symbols["decimal_split"])[0]) {
					if dot && rune(input[i]) == []rune(config.Symbols["decimal_split"])[0] {
						cfmt.Printf("{{Error:}}::red|bold Unable to parse number at %v, reading multiple decimal splits.\n", i)
						return nil, errors.New("multiple decimal splits")
					} else if rune(input[i]) == []rune(config.Symbols["decimal_split"])[0] {
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
				tokens = append(tokens, Token{NUMBER, num, ""})
			} else if unicode.IsSpace(rune(input[i])) {
				// Skip empty space
				i += 1
			} else {
				// Search for constants or variables 
				str := ""

				for i < len(input) && unicode.IsLetter(rune(input[i])) {
					str += string(input[i])
					i += 1
				}

				// Check Constants
				if val, ok := config.Constants[str]; ok {
					tokens = append(tokens, Token{NUMBER, val, ""})
				} else if str == "sqrt" { 
					tokens = append(tokens, Token{SQRT, 0.0, ""})
				} else {
					j := 0
					for j < len(str) {
						
						// Debug
						if config.Options["show_debug_process"] {
							if _, ok := variables[string(str[j])]; ok {
								cfmt.Printf("{{Notice:}}::blue|bold found defined variable %s with value.\n", string(str[j]))
							} else {
								cfmt.Printf("{{Notice:}}::blue|bold found undefined variable %s.\n", string(str[j]))
							}
						}
						variableOccurrence = append(variableOccurrence, string(str[j]))
						tokens = append(tokens, Token{VARIABLE, 0.0, string(str[j])})
						j += 1
					}
					// if i >= len(input) {
					// 	i -= 1
					// }
					// cfmt.Printf("{{Error:}}::red|bold Unable to parse symbole at %v, reading unrecognised character: '%s'.\n", i, string(input[i]))
					// return nil, errors.New("unrecognised character")
				}
			}
		}
	}
	return tokens, nil
}

