package nodes

import (
	"bytes"
	"guts/parser/expressions"
)

type LoopBlock interface {
	Node
	IndexKey() string
	ValueKey() string
	ItemsKey() string
	ExpressionType() expressions.ExpressionType
}

type loopBlock struct {
	node
	indexKey string
	valueKey string
	itemsKey string
	typ      expressions.ExpressionType
}

func NewLoopBlock(indexKey, valueKey, itemsKey string, typ expressions.ExpressionType) LoopBlock {
	return &loopBlock{
		node: node{
			name: "#loop",
		},
		indexKey: indexKey,
		valueKey: valueKey,
		itemsKey: itemsKey,
		typ:      typ,
	}
}

func (e *loopBlock) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteString("{for ")
	buf.WriteString(e.indexKey)
	buf.WriteString(", ")
	buf.WriteString(e.valueKey)
	buf.WriteString(" in ")
	buf.WriteString(e.itemsKey)
	if e.typ != nil {
		buf.WriteString(":")
		buf.WriteString(e.typ.String())
	}
	buf.WriteString("}")

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	buf.WriteString("{/for}")
	return buf.String()
}

func (e *loopBlock) TextContent() string {
	return ""
}

func (e *loopBlock) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\"name\": \"#loop\", \"indexKey\": \"")
	buf.WriteString(e.indexKey)
	buf.WriteString("\", \"valueKey\": \"")
	buf.WriteString(e.valueKey)
	buf.WriteString("\", \"itemsKey\": \"")
	buf.WriteString(e.itemsKey)
	buf.WriteString("\", \"typ\": \"")
	if e.typ != nil {
		buf.WriteString(e.typ.String())
	}
	buf.WriteString("\", \"children\": [")
	for _, child := range e.children {
		buf.WriteString(child.String())
	}
	buf.WriteString("]}")
	return buf.String()
}

func (e *loopBlock) IndexKey() string {
	return e.indexKey
}

func (e *loopBlock) ValueKey() string {
	return e.valueKey
}

func (e *loopBlock) ItemsKey() string {
	return e.itemsKey
}

func (e *loopBlock) ExpressionType() expressions.ExpressionType {
	return e.typ
}
