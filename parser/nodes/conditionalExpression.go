package nodes

import (
	"bytes"
)

type ConditionalExpression interface {
	Node
	SetCondition(string)
	Condition() string
	SetNext(ConditionalExpression)
	Next() ConditionalExpression
}

type conditionalExpression struct {
	node
	condition string
	next      ConditionalExpression
}

func NewConditionalExpression() ConditionalExpression {
	return &conditionalExpression{}
}

func (e *conditionalExpression) SetCondition(condition string) {
	e.condition = condition
}

func (e *conditionalExpression) Condition() string {
	return e.condition
}

func (e *conditionalExpression) SetNext(next ConditionalExpression) {
	next.setParent(e.Parent())
	e.next = next
}

func (e *conditionalExpression) Next() ConditionalExpression {
	return e.next
}

func (e *conditionalExpression) OuterHTML() string {
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

func (e *conditionalExpression) String() string {
	var nextStr string
	if e.next != nil {
		nextStr = e.next.String()
	}
	return "{\"name\": \"#conditional-expression\", \"condition\": \"" + e.condition + "\", \"next\": " + nextStr + "}"
}
