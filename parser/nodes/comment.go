package nodes

import "encoding/json"

type Comment interface {
	Node
}

type comment struct {
	node
	comment string
}

func NewComment(text string) TextNode {
	return &comment{
		node: node{
			name: "#comment",
		},
		comment: text,
	}
}

func (t *comment) TextContent() string {
	return ""
}

func (t *comment) OuterHTML() string {
	return "<!--" + t.comment + "-->"
}

func (t *comment) Children() []Node {
	return []Node{}
}

func (t *comment) Append(children ...Node) {
	// no op
}

func (t *comment) String() string {
	text, _ := json.Marshal(t.comment)
	return "{\"name\": \"#comment\", \"comment\": " + string(text) + "}"
}
