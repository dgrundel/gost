package nodes

import (
	"bytes"
	"strings"
)

type Node interface {
	Name() string
	TextContent() string
	OuterHTML() string
	Parent() Node
	setParent(Node)
	Children() []Node
	Append(children ...Node)
	String() string
}

type node struct {
	name     string
	parent   Node
	children []Node
}

func NewNode(name string) Node {
	return &node{name: name}
}

func (t *node) Name() string {
	return t.name
}

func (t *node) TextContent() string {
	var builder strings.Builder
	for _, child := range t.children {
		builder.WriteString(child.TextContent())
	}
	return builder.String()
}

func (t *node) OuterHTML() string {
	var buf bytes.Buffer
	for _, child := range t.children {
		buf.WriteString(child.OuterHTML())
	}
	return buf.String()
}

func (t *node) Parent() Node {
	return t.parent
}

func (t *node) setParent(n Node) {
	t.parent = n
}

func (t *node) Children() []Node {
	return t.children
}

func (t *node) Append(children ...Node) {
	t.children = append(t.children, children...)
	for _, c := range children {
		c.setParent(t)
	}
}

func (t *node) String() string {
	var fields = []string{
		"\"name\": \"" + t.Name() + "\"",
	}

	if len(t.Children()) == 0 {
		fields = append(fields, "\"textContent\": \""+t.TextContent()+"\"")
	} else {
		var children []string
		for _, c := range t.Children() {
			children = append(children, c.String())
		}
		fields = append(fields, "\"children\": ["+strings.Join(children, ", ")+"]")
	}

	return "{" + strings.Join(fields, ", ") + "}"
}
