package shared

const (
	NUMBER       = iota
	PLUS         = iota
	MINUS        = iota
	MULTIPLY     = iota
	DIVIDE       = iota
	POWER        = iota
	SQRT         = iota
	LPARENTHESES = iota
	RPARENTHESES = iota
	EQUAL        = iota
	VARIABLE     = iota
)

func GetDefualtConfig() Config {
	return Config{
		Options: map[string]bool{
			"nerdfont":           true,
			"show_debug_process": false,
		},
		Symbols: map[string]string{
			"decimal_split": ".",
			"plus":          "+",
			"minus":         "-",
			"multiply":      "*",
			"divide":        "/",
			"power":         "^",
			"sqrt":          "sqrt",
			"l_parentheses": "(",
			"r_parentheses": ")",
			"equal":         "=",
		},
		Constants: map[string]float64{
			"pi":  3.14159265358979323846264338327950288419716939937510582097494459,
			"phi": 1.61803398874989484820458683436563811772030917980576286213544862,
			"e":   2.71828182845904523536028747135266249775724709369995957496696763,
		},
	}
}
