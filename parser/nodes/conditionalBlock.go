package nodes

import (
	"bytes"
	"gost/parser/expressions"
)

type ConditionalBlock interface {
	Node
	SetCondition(expressions.BooleanExpression)
	Condition() expressions.BooleanExpression
	SetNext(ConditionalBlock)
	Next() ConditionalBlock
	IsConditionalBlock() bool
}

type conditionalBlock struct {
	node
	condition expressions.BooleanExpression
	next      ConditionalBlock
}

func NewConditionalBlock() ConditionalBlock {
	return &conditionalBlock{}
}

func (e *conditionalBlock) SetCondition(condition expressions.BooleanExpression) {
	e.condition = condition
}

func (e *conditionalBlock) Condition() expressions.BooleanExpression {
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
	buf.WriteString(e.condition.String())
	buf.WriteString("}")

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	next := e.next
	for next != nil {
		if next.Condition() != nil {
			buf.WriteString("{else if ")
			buf.WriteString(next.Condition().String())
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
	var condStr string
	if e.condition != nil {
		condStr = e.condition.String()
	}
	return "{\"name\": \"#conditional-block\", \"condition\": \"" + condStr + "\", \"next\": " + nextStr + "}"
}

func (e *conditionalBlock) IsConditionalBlock() bool {
	return true
}
