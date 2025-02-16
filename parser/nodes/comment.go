package nodes

import "encoding/json"

type Comment interface {
	Node
}

type comment struct {
	node
	textContent string
}

func NewComment(text string) TextNode {
	return &comment{
		node: node{
			name: "#comment",
		},
		textContent: text,
	}
}

func (t *comment) TextContent() string {
	return ""
}

func (t *comment) OuterHTML() string {
	return "<!--" + t.textContent + "-->"
}

func (t *comment) Children() []Node {
	return []Node{}
}

func (t *comment) Append(children ...Node) {
	// no op
}

func (t *comment) String() string {
	text, _ := json.Marshal(t.textContent)
	return "{\"name\": \"#comment\", \"comment\": " + string(text) + "}"
}
