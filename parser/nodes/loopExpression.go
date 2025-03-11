package nodes

import (
	"bytes"
)

type LoopExpression interface {
	Node
}

type loopExpression struct {
	node
	indexKey string
	valueKey string
	itemsKey string
	typ      string
}

func NewLoopExpression(indexKey, valueKey, itemsKey, typ string) LoopExpression {
	return &loopExpression{
		indexKey: indexKey,
		valueKey: valueKey,
		itemsKey: itemsKey,
		typ:      typ,
	}
}

func (e *loopExpression) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteString("{for ")
	buf.WriteString(e.indexKey)
	buf.WriteString(", ")
	buf.WriteString(e.valueKey)
	buf.WriteString(" in ")
	buf.WriteString(e.itemsKey)
	if e.typ != "" {
		buf.WriteString(":")
		buf.WriteString(e.typ)
	}
	buf.WriteString("}")

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	buf.WriteString("{/for}")
	return buf.String()
}

func (e *loopExpression) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\"name\": \"#loop-expression\", \"indexKey\": \"")
	buf.WriteString(e.indexKey)
	buf.WriteString("\", \"valueKey\": \"")
	buf.WriteString(e.valueKey)
	buf.WriteString("\", \"itemsKey\": \"")
	buf.WriteString(e.itemsKey)
	buf.WriteString("\", \"typ\": \"")
	buf.WriteString(e.typ)
	buf.WriteString("\", \"children\": [")
	for _, child := range e.children {
		buf.WriteString(child.String())
	}
	buf.WriteString("]}")
	return buf.String()
}
