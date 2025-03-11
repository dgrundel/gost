package nodes

import (
	"bytes"
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
	GetAttribute(string) AttributeValue
	SetAttribute(string, AttributeValue)
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
	var buf bytes.Buffer
	buf.WriteByte('<')
	buf.WriteString(t.name)

	t.attributes.Iterator()(func(key string, value AttributeValue) bool {
		buf.WriteByte(' ')
		buf.WriteString(key)

		if value != nil && !value.IsEmpty() {
			buf.WriteString("=")
			buf.WriteString(value.OuterHTML())
		}

		return true
	})

	buf.WriteByte('>')

	if t.void {
		return buf.String()
	}

	for _, child := range t.children {
		buf.WriteString(child.OuterHTML())
	}

	buf.WriteString("</")
	buf.WriteString(t.name)
	buf.WriteByte('>')

	return buf.String()
}

func (t *element) GetAttribute(name string) AttributeValue {
	return t.attributes.GetAttribute(name)
}

func (t *element) SetAttribute(name string, value AttributeValue) {
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
