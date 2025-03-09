package nodes

type Document interface {
	Node
	GetDeclaredTypes() map[string]ExpressionType
}

type document struct {
	node
	declaredTypes map[string]ExpressionType
}

func NewDocument() Document {
	return &document{
		node:          node{name: "#document"},
		declaredTypes: make(map[string]ExpressionType),
	}
}

func (t *document) GetDeclaredTypes() map[string]ExpressionType {
	return t.declaredTypes
}

func (t *document) Parent() Node {
	return nil
}
