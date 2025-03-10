package nodes

import (
	"bytes"
	"io"
)

type ConditionalExpression interface {
	Node
	SetCondition(Condition)
	SetPrev(ConditionalExpression)
	SetNext(ConditionalExpression)
}

type conditionalExpression struct {
	node
	condition Condition
	prev      ConditionalExpression
	next      ConditionalExpression
}

func NewConditionalExpression() ConditionalExpression {
	return &conditionalExpression{}
}

func (e *conditionalExpression) SetCondition(condition Condition) {
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

	if e.prev == nil && e.condition != nil {
		buf.WriteString("{if ")
		buf.WriteString(e.condition.OuterHTML())
		buf.WriteString("} ")
	} else if e.prev != nil && e.condition != nil {
		buf.WriteString("{else if ")
		buf.WriteString(e.condition.OuterHTML())
		buf.WriteString("} ")
	} else if e.prev != nil && e.condition == nil {
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

func (e *conditionalExpression) Render(c RenderContext, w io.Writer) error {
	// nil condition expected in else branch
	if e.condition == nil || e.condition.Evaluate(c) {
		for _, child := range e.children {
			err := child.Render(c, w)
			if err != nil {
				return err
			}
		}
	} else if e.next != nil { // else branch
		return e.next.Render(c, w)
	}
	return nil
}

type Condition interface {
	Evaluate(RenderContext) bool
	OuterHTML() string
}

type stringCondition struct {
	expression string
}

func NewStringCondition(expression string) Condition {
	return &stringCondition{expression: expression}
}

func (s *stringCondition) Evaluate(c RenderContext) bool {
	return true
}

func (s *stringCondition) OuterHTML() string {
	return s.expression
}
