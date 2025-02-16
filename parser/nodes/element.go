package nodes

import (
	"encoding/json"
	"strings"
)

var _voidElements = map[string]bool{
	"!doctype": true, // hack
	"area":     true,
	"base":     true,
	"br":       true,
	"col":      true,
	"embed":    true,
	"hr":       true,
	"img":      true,
	"input":    true,
	"link":     true,
	"meta":     true,
	"param":    true,
	"source":   true,
	"track":    true,
	"wbr":      true,
}

type Element interface {
	Node
	IsVoid() bool
	GetAttribute(string) string
	SetAttribute(string, string)
}

type element struct {
	node
	void       bool
	attributes Attributes
}

func NewElement(name string, void bool) Element {
	void = void || _voidElements[name]

	return &element{
		node:       node{name: name},
		void:       void,
		attributes: NewAttributes(),
	}
}

func (t *element) IsVoid() bool {
	return t.void
}

func (t *element) OuterHTML() string {
	var builder strings.Builder
	builder.WriteByte('<')
	builder.WriteString(t.name)

	t.attributes.Iterator()(func(key, value string) bool {
		builder.WriteByte(' ')
		builder.WriteString(key)

		if value != "" {
			builder.WriteString("=\"")
			builder.WriteString(value)
			builder.WriteByte('"')
		}

		return true
	})

	builder.WriteByte('>')

	if t.void {
		return builder.String()
	}

	for _, child := range t.children {
		builder.WriteString(child.OuterHTML())
	}

	builder.WriteString("</")
	builder.WriteString(t.name)
	builder.WriteByte('>')
	return builder.String()
}

func (t *element) GetAttribute(name string) string {
	return t.attributes.GetAttribute(name)
}

func (t *element) SetAttribute(name string, value string) {
	t.attributes.SetAttribute(name, value)
}

func (t *element) String() string {
	attrs, _ := json.Marshal(t.attributes.All())

	var fields = []string{
		"\"name\": \"" + t.Name() + "\"",
		"\"attrs\": " + string(attrs),
	}

	if len(t.Children()) > 0 {
		var children []string
		for _, c := range t.Children() {
			children = append(children, c.String())
		}
		fields = append(fields, "\"children\": ["+strings.Join(children, ", ")+"]")
	}

	return "{" + strings.Join(fields, ", ") + "}"
}
