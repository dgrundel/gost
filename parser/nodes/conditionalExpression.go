package nodes

import (
	"bytes"
)

type ConditionalExpression interface {
	Node
	SetCondition(string)
	SetPrev(ConditionalExpression)
	SetNext(ConditionalExpression)
}

type conditionalExpression struct {
	node
	condition string
	prev      ConditionalExpression
	next      ConditionalExpression
}

func NewConditionalExpression() ConditionalExpression {
	return &conditionalExpression{}
}

func (e *conditionalExpression) SetCondition(condition string) {
	e.condition = condition
}

func (e *conditionalExpression) SetPrev(prev ConditionalExpression) {
	e.prev = prev
}

func (e *conditionalExpression) SetNext(next ConditionalExpression) {
	e.next = next
}

func (e *conditionalExpression) OuterHTML() string {
	var buf bytes.Buffer

	if e.prev == nil && e.condition != "" {
		buf.WriteString("{if ")
		buf.WriteString(e.condition)
		buf.WriteString("} ")
	} else if e.prev != nil && e.condition != "" {
		buf.WriteString("{else if ")
		buf.WriteString(e.condition)
		buf.WriteString("} ")
	} else if e.prev != nil && e.condition == "" {
		buf.WriteString("{else} ")
	} else {
		buf.WriteString("{if true} ")
	}

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	buf.WriteString("{/if}")
	return buf.String()
}
