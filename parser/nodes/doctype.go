package nodes

import "fmt"

type Doctype interface {
	Node
}

type doctype struct {
	node
	textContent string
}

func NewDoctype(text string) Doctype {
	return &doctype{textContent: text}
}

func (t *doctype) Name() string {
	return "#doctype"
}

func (t *doctype) TextContent() string {
	return t.textContent
}

func (t *doctype) Children() []Node {
	return []Node{}
}

func (t *doctype) Append(children ...Node) {
	// no op
}

func (t *doctype) String() string {
	return fmt.Sprintf("\"%s\"", t.Name())
}
