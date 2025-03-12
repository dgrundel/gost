package attributes

import (
	"fmt"
	"gost/parser/expressions"
	"strings"
)

type AttributeValueSpread interface {
	AttributeValue
	Key() string
	ExpressionType() expressions.ExpressionType
}

func NewAttributeValueSpread(s string) (AttributeValueSpread, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty spread attribute")
	}
	if !strings.HasPrefix(s, "...") {
		return nil, fmt.Errorf("invalid spread attribute: %s", s)
	}
	s = strings.TrimPrefix(s, "...")
	key, expressionType, err := splitType(s)
	if err != nil {
		return nil, err
	}
	return &attributeValueSpread{
		key:            key,
		expressionType: expressionType,
	}, nil
}

type attributeValueSpread struct {
	expressionType expressions.ExpressionType
	key            string
}

func (s *attributeValueSpread) OuterHTML() string {
	if s.expressionType == nil {
		return "{..." + s.key + "}"
	}
	return "{..." + s.key + ":" + s.expressionType.String() + "}"
}

func (s *attributeValueSpread) IsEmpty() bool {
	return s.key == ""
}

func (s *attributeValueSpread) Key() string {
	return s.key
}

func (s *attributeValueSpread) ExpressionType() expressions.ExpressionType {
	return s.expressionType
}
