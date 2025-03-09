package nodes

import "io"

type ConditionalExpression interface {
	Node
}

type conditionalExpression struct {
	node
	condition func(RenderContext) bool
	next      ConditionalExpression
}

func NewConditionalExpression() ConditionalExpression {
	return &conditionalExpression{}
}

func (e *conditionalExpression) Render(c RenderContext, w io.Writer) error {
	// nil condition expected in else branch
	if e.condition == nil || e.condition(c) {
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
