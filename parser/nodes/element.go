package nodes

import (
	"bytes"
	"gost/parser/nodes/attributes"
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
	Attributes() attributes.Attributes
	SetBind(bind string)
	Bind() string
}

type element struct {
	node
	void       bool
	attributes attributes.Attributes
	bind       string
}

func NewElement(name string, void bool) Element {
	void = void || _voidElements[name]

	return &element{
		node:       node{name: name},
		void:       void,
		attributes: attributes.NewAttributes(),
	}
}

func (t *element) IsVoid() bool {
	return t.void
}

func (t *element) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteByte('<')
	buf.WriteString(t.name)

	t.attributes.Iterator()(func(key string, value attributes.AttributeValue) bool {
		buf.WriteByte(' ')
		buf.WriteString(key)

		if value != nil && !value.IsEmpty() {
			buf.WriteString("=")
			buf.WriteString(value.OuterHTML())
		}

		return true
	})

	spread := t.attributes.GetSpreadAttribute()
	if spread != nil && !spread.IsEmpty() {
		buf.WriteByte(' ')
		buf.WriteString(spread.OuterHTML())
	}

	if t.bind != "" {
		buf.WriteString(" data-bind-id=\"")
		buf.WriteString(t.bind)
		buf.WriteByte('"')
	}

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

func (t *element) Attributes() attributes.Attributes {
	return t.attributes
}

func (t *element) SetBind(bind string) {
	t.bind = bind
}

func (t *element) Bind() string {
	return t.bind
}

func (t *element) String() string {
	var fields = []string{
		"\"name\": \"" + t.Name() + "\"",
		"\"attrs\": " + t.attributes.String(),
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
