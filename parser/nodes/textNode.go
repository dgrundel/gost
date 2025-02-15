package nodes

type TextNode interface {
	Node
}

type textNode struct {
	node
	textContent string
}

func NewTextNode(text string) TextNode {
	return &textNode{textContent: text}
}

func (t *textNode) Name() string {
	return "#text"
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
