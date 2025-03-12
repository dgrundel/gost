package nodes

import (
	"bytes"
	"fmt"
	"gost/parser/expressions"
	"strings"
)

type Iter = func(yield func(key string, value AttributeValue) bool)

type AttributeValue interface {
	OuterHTML() string
	IsEmpty() bool
}

type AttributeValueString string

func (s AttributeValueString) OuterHTML() string {
	return "\"" + string(s) + "\""
}

func (s AttributeValueString) IsEmpty() bool {
	return s == ""
}

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

type Attributes interface {
	GetAttribute(key string) AttributeValue
	SetAttribute(key string, value AttributeValue)
	GetSpreadAttribute() AttributeValueSpread
	SetSpreadAttribute(value AttributeValueSpread)
	Iterator() Iter
	All() map[string]AttributeValue
	String() string
}

type attrs struct {
	keys   []string
	values map[string]AttributeValue
	spread AttributeValueSpread
}

func NewAttributes() Attributes {
	return &attrs{
		values: make(map[string]AttributeValue),
	}
}

func (a *attrs) GetAttribute(key string) AttributeValue {
	return a.values[key]
}

func (a *attrs) SetAttribute(key string, value AttributeValue) {
	_, exists := a.values[key]
	if !exists {
		a.keys = append(a.keys, key)
	}

	a.values[key] = value
}

func (a *attrs) GetSpreadAttribute() AttributeValueSpread {
	return a.spread
}

func (a *attrs) SetSpreadAttribute(value AttributeValueSpread) {
	a.spread = value
}

func (a *attrs) Iterator() Iter {
	return func(yield func(key string, value AttributeValue) bool) {
		for _, key := range a.keys {
			if !yield(key, a.values[key]) {
				return
			}
		}
	}
}

func (a *attrs) All() map[string]AttributeValue {
	return a.values
}

func (a *attrs) String() string {
	var buf bytes.Buffer

	for _, key := range a.keys {
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(a.values[key].OuterHTML())
	}

	if a.spread != nil && !a.spread.IsEmpty() {
		buf.WriteString("...:")
		buf.WriteString(a.spread.OuterHTML())
	}

	return buf.String()
}

func splitType(s string) (string, expressions.ExpressionType, error) {
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		expressionType, ok := expressions.ParseExpressionType(parts[1])
		if !ok {
			return "", nil, fmt.Errorf("invalid expression type: %s", parts[1])
		}
		return strings.TrimSpace(parts[0]), expressionType, nil
	}
	return strings.TrimSpace(s), nil, nil
}
