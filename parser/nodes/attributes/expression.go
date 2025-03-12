package attributes

import "gost/parser/expressions"

type AttributeValueExpression interface {
	AttributeValue
	Key() string
	ExpressionType() expressions.ExpressionType
}

func NewAttributeValueExpression(s string) (AttributeValueExpression, error) {
	key, expressionType, err := splitType(s)
	if err != nil {
		return nil, err
	}
	return &attributeValueExpression{
		key:            key,
		expressionType: expressionType,
	}, nil
}

type attributeValueExpression struct {
	expressionType expressions.ExpressionType
	key            string
}

func (e *attributeValueExpression) OuterHTML() string {
	if e.expressionType == nil {
		return "{" + e.key + "}"
	}
	return "{" + e.key + ":" + e.expressionType.String() + "}"
}

func (e *attributeValueExpression) IsEmpty() bool {
	return e.key == ""
}

func (e *attributeValueExpression) Key() string {
	return e.key
}

func (e *attributeValueExpression) ExpressionType() expressions.ExpressionType {
	return e.expressionType
}
