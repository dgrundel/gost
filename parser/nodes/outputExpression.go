package nodes

import "bytes"

type OutputExpression interface {
	Node
}

type outputExpression struct {
	node
	key string
	typ string
}

func NewOutputExpression(key string, typ string) OutputExpression {
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
	if o.typ != "" {
		buf.WriteString(":")
		buf.WriteString(o.typ)
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
	return "{\"name\": \"#output-expression\", \"key\": " + string(o.key) + "}"
}
