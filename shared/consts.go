package shared

const (
	NUMBER       = iota // 0
	PLUS         = iota // 1
	MINUS        = iota // 2
	MULTIPLY     = iota // 3
	DIVIDE       = iota // 4
	POWER        = iota // 5
	SQRT         = iota // 6
	LPARENTHESES = iota // 7
	RPARENTHESES = iota // 8
	EQUAL        = iota // 9
	VARIABLE     = iota // 10
	COMMA        = iota // 11
	FUNCTION     = iota // 12
)

func GetDefualtConfig() Config {
	return Config{
		Version: "0.1.5",
		Options: map[string]bool{
			"nerdfont":           true,
			"show_debug_process": false,
		},
		Symbols: map[string]string{
			"decimal_split":   ".",
			"parameter_split": ",",
			"plus":            "+",
			"minus":           "-",
			"multiply":        "*",
			"divide":          "/",
			"power":           "^",
			"sqrt":            "sqrt",
			"l_parentheses":   "(",
			"r_parentheses":   ")",
			"equal":           "=",
		},
		Constants: map[string]float64{
			"pi":  3.14159265358979323846264338327950288419716939937510582097494459,
			"phi": 1.61803398874989484820458683436563811772030917980576286213544862,
			"e":   2.71828182845904523536028747135266249775724709369995957496696763,
		},
	}
}
