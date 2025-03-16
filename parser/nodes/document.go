package nodes

import (
	"bytes"
	"fmt"
	"guts/parser/expressions"
	"strings"
)

type Document interface {
	Node
	GetDeclaredTypes() map[string]expressions.ExpressionType
	AddDeclaredType(name string, expressionType expressions.ExpressionType) error
}

type document struct {
	node
	declaredTypes map[string]expressions.ExpressionType
}

func NewDocument() Document {
	return &document{
		node:          node{name: "#document"},
		declaredTypes: make(map[string]expressions.ExpressionType),
	}
}

func (t *document) GetDeclaredTypes() map[string]expressions.ExpressionType {
	return t.declaredTypes
}

func (t *document) AddDeclaredType(name string, expressionType expressions.ExpressionType) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty type name")
	}
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
