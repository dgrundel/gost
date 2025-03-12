package nodes

import (
	"bytes"
)

type ConditionalBlock interface {
	Node
	SetCondition(string)
	Condition() string
	SetNext(ConditionalBlock)
	Next() ConditionalBlock
}

type conditionalBlock struct {
	node
	condition string // TODO: make this an expression
	next      ConditionalBlock
}

func NewConditionalBlock() ConditionalBlock {
	return &conditionalBlock{}
}

func (e *conditionalBlock) SetCondition(condition string) {
	e.condition = condition
}

func (e *conditionalBlock) Condition() string {
	return e.condition
}

func (e *conditionalBlock) SetNext(next ConditionalBlock) {
	next.setParent(e.Parent())
	e.next = next
}

func (e *conditionalBlock) Next() ConditionalBlock {
	return e.next
}

func (e *conditionalBlock) OuterHTML() string {
	var buf bytes.Buffer

	buf.WriteString("{if ")
	buf.WriteString(e.condition)
	buf.WriteString("}")

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	next := e.next
	for next != nil {
		if next.Condition() != "" {
			buf.WriteString("{else if ")
			buf.WriteString(next.Condition())
			buf.WriteString("}")
		} else {
			buf.WriteString("{else}")
		}

		for _, child := range next.Children() {
			buf.WriteString(child.OuterHTML())
		}

		next = next.Next()
	}

	buf.WriteString("{/if}")

	return buf.String()
}

func (e *conditionalBlock) String() string {
	var nextStr string
	if e.next != nil {
		nextStr = e.next.String()
	}
	return "{\"name\": \"#conditional-block\", \"condition\": \"" + e.condition + "\", \"next\": " + nextStr + "}"
}
