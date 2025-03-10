package nodes

import (
	"bytes"
	"errors"
	"io"
)

type LoopExpression interface {
	Node
}

type loopExpression struct {
	node
	indexKey string
	valueKey string
	itemsKey string
	typ      string
}

func NewLoopExpression(indexKey, valueKey, itemsKey, typ string) LoopExpression {
	return &loopExpression{
		indexKey: indexKey,
		valueKey: valueKey,
		itemsKey: itemsKey,
		typ:      typ,
	}
}

func (e *loopExpression) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteString("{for ")
	buf.WriteString(e.indexKey)
	buf.WriteString(", ")
	buf.WriteString(e.valueKey)
	buf.WriteString(" in ")
	buf.WriteString(e.itemsKey)
	buf.WriteString("}")

	for _, child := range e.children {
		buf.WriteString(child.OuterHTML())
	}

	buf.WriteString("{/for}")
	return buf.String()
}

func (e *loopExpression) Render(c RenderContext, w io.Writer) error {
	items, ok := c.Get(e.itemsKey)
	if !ok {
		return errors.New("items not found")
	}

	slice, ok := items.([]any)
	if !ok {
		return errors.New("items is not a slice")
	}

	for i, item := range slice {
		// Create a new context with loop variables
		loopContext := c.WithData(map[string]any{
			e.indexKey: i,
			e.valueKey: item,
		})

		// Render each child with the loop context
		for _, child := range e.children {
			err := child.Render(loopContext, w)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
