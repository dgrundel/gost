package nodes

import "fmt"

type Comment interface {
	Node
}

type comment struct {
	node
	textContent string
}

func NewComment(text string) Comment {
	return &comment{textContent: text}
}

func (t *comment) Name() string {
	return "#comment"
}

func (t *comment) TextContent() string {
	return t.textContent
}

func (t *comment) Children() []Node {
	return []Node{}
}

func (t *comment) Append(children ...Node) {
	// no op
}

func (t *comment) String() string {
	return fmt.Sprintf("\"%s\"", t.Name())
}
