package nodes

type Document interface {
	Node
}

type document struct {
	node
}

func NewDocument() Document {
	return &document{
		node: node{name: "#document"},
	}
}

func (t *document) Parent() Node {
	return nil
}
