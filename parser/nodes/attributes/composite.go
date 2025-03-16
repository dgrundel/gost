package attributes

import (
	"bytes"
	"guts/parser/expressions"
)

type AttributeValueComposite interface {
	AttributeValue
	DeclaredTypes() map[string]expressions.ExpressionType
	Values() []AttributeValue
}

type attributeValueComposite struct {
	values        []AttributeValue
	declaredTypes map[string]expressions.ExpressionType
}

func NewAttributeValueComposite(s string) (AttributeValueComposite, error) {
	values := make([]AttributeValue, 0, 5)
	declaredTypes := make(map[string]expressions.ExpressionType)

	var buf bytes.Buffer
	var inExpression bool
	for _, c := range s {
		switch c {
		case '{':
			if buf.Len() > 0 {
				values = append(values, AttributeValueString(buf.String()))
				buf.Reset()
			}
			inExpression = true
		case '}':
			if buf.Len() > 0 {
				if inExpression {
					expr, err := NewAttributeValueExpression(buf.String())
					if err != nil {
						return nil, err
					}
					if expr.ExpressionType() != nil {
						declaredTypes[expr.Key()] = expr.ExpressionType()
					}
					values = append(values, expr)
					buf.Reset()
				} else {
					buf.WriteRune(c)
				}
			}
			inExpression = false
		default:
			buf.WriteRune(c)
		}
	}

	if buf.Len() > 0 {
		if inExpression {
			values = append(values, AttributeValueString("{"+buf.String()))
		} else {
			values = append(values, AttributeValueString(buf.String()))
		}
	}

	return &attributeValueComposite{values: values, declaredTypes: declaredTypes}, nil
}

func (c *attributeValueComposite) OuterHTML() string {
	var buf bytes.Buffer
	buf.WriteRune('"')
	for _, v := range c.values {
		buf.WriteString(v.OuterHTML())
	}
	buf.WriteRune('"')
	return buf.String()
}

func (c *attributeValueComposite) IsEmpty() bool {
	return len(c.values) == 0
}

func (c *attributeValueComposite) DeclaredTypes() map[string]expressions.ExpressionType {
	return c.declaredTypes
}

func (c *attributeValueComposite) Values() []AttributeValue {
	return c.values
}
