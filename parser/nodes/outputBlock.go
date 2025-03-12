package nodes

import (
	"bytes"
	"gost/parser/expressions"
)

type OutputBlock interface {
	Node
	Key() string
	ExpressionType() expressions.ExpressionType
}

type outputBlock struct {
	node
	key string
	typ expressions.ExpressionType
}

func NewOutputExpression(key string, typ expressions.ExpressionType) OutputBlock {
	return &outputBlock{
		node: node{
			name: "#output-expression",
		},
		key: key,
		typ: typ,
	}
}

func (o *outputBlock) TextContent() string {
	return ""
}

func (o *outputBlock) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteString(o.key)
	if o.typ != nil {
		buf.WriteString(":")
		buf.WriteString(o.typ.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func (o *outputBlock) Children() []Node {
	return []Node{}
}

func (o *outputBlock) Append(children ...Node) {
	// no op
}

func (o *outputBlock) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\"name\": \"#output-expression\", \"key\": \"")
	buf.WriteString(o.key)
	buf.WriteString("\", \"typ\": \"")
	if o.typ != nil {
		buf.WriteString(o.typ.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func (o *outputBlock) Key() string {
	return o.key
}

func (o *outputBlock) ExpressionType() expressions.ExpressionType {
	return o.typ
}
