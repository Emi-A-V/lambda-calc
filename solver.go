package lambdaengine

func solve(node *Node) (*Node, error) {
	simp, err := simplify(node, SOLVE)
	if err != nil {
		return nil, err
	}
	printTree(simp)

	return simp, nil
}
