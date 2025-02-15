package nodes

import "strings"

type Document interface {
	Node
	OuterHTML() string
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

func (t *document) OuterHTML() string {
	var builder strings.Builder
	for _, child := range t.children {
		el, ok := child.(Element)
		if ok {
			builder.WriteString(el.OuterHTML())
		}
	}
	return builder.String()
}
