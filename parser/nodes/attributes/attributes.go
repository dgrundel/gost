package attributes

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
