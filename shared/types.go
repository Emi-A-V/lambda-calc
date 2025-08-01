package shared

type Config struct {
	Version   string
	Options   map[string]bool
	Symbols   map[string]string
	Constants map[string]float64
}

type Token struct {
	TokenType int
	Value     float64
	Variable  string
}

type Node struct {
	OperationType int
	Value         float64
	Variable      string
	LNode         *Node
	RNode         *Node
	Associative   []*Node
}

type Function struct {
	Parameters []*Node
	Equation   *Node
}
