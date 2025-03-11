package nodes

import (
	"bytes"
	"fmt"
)

type Document interface {
	Node
	GetDeclaredTypes() map[string]ExpressionType
	AddDeclaredType(name string, expressionType ExpressionType) error
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

func (t *document) AddDeclaredType(name string, expressionType ExpressionType) error {
	if declared, ok := t.declaredTypes[name]; ok && !declared.Equals(expressionType) {
		return fmt.Errorf("%s is already declared with different type: %s", name, declared.String())
	}
	t.declaredTypes[name] = expressionType
	return nil
}

func (t *document) Parent() Node {
	return nil
}

func (t *document) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\"name\": \"#document\", \"declaredTypes\": [")
	for name, expressionType := range t.declaredTypes {
		buf.WriteString(fmt.Sprintf("{\"name\": \"%s\", \"expressionType\": \"%s\"}, ", name, expressionType.String()))
	}
	buf.WriteString("], \"children\": [")
	for _, child := range t.children {
		buf.WriteString(child.String())
		buf.WriteString(", ")
	}
	buf.WriteString("]}")
	return buf.String()
}
