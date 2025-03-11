package nodes

import (
	"encoding/json"
)

type TextNode interface {
	Node
}

type textNode struct {
	node
	textContent string
}

func NewTextNode(text string) TextNode {
	return &textNode{
		node: node{
			name: "#text",
		},
		textContent: text,
	}
}

func (t *textNode) TextContent() string {
	return t.textContent
}

func (t *textNode) OuterHTML() string {
	return t.textContent
}

func (t *textNode) Children() []Node {
	return []Node{}
}

func (t *textNode) Append(children ...Node) {
	// no op
}

func (t *textNode) String() string {
	text, _ := json.Marshal(t.textContent)
	return "{\"name\": \"#text\", \"textContent\": " + string(text) + "}"
}
