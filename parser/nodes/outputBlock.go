package nodes

import (
	"bytes"
	"gost/parser/expressions"
)

type OutputExpression interface {
	Node
}

type outputExpression struct {
	node
	key string
	typ expressions.ExpressionType
}

func NewOutputExpression(key string, typ expressions.ExpressionType) OutputExpression {
	return &outputExpression{
		node: node{
			name: "#output-expression",
		},
		key: key,
		typ: typ,
	}
}

func (o *outputExpression) TextContent() string {
	return ""
}

func (o *outputExpression) OuterHTML() string {
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

func (o *outputExpression) Children() []Node {
	return []Node{}
}

func (o *outputExpression) Append(children ...Node) {
	// no op
}

func (o *outputExpression) String() string {
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
