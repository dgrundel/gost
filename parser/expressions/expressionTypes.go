package expressions

import (
	"regexp"
	"strings"
)

type ExpressionBaseType string

const (
	ExpressionBaseTypeString ExpressionBaseType = "string"
	ExpressionBaseTypeInt    ExpressionBaseType = "int"
	ExpressionBaseTypeFloat  ExpressionBaseType = "float"
	ExpressionBaseTypeBool   ExpressionBaseType = "bool"
	ExpressionBaseTypeArray  ExpressionBaseType = "array"
	ExpressionBaseTypeMap    ExpressionBaseType = "map"
)

var expressionBaseTypeMap = map[string]ExpressionBaseType{
	"string": ExpressionBaseTypeString,
	"int":    ExpressionBaseTypeInt,
	"float":  ExpressionBaseTypeFloat,
	"bool":   ExpressionBaseTypeBool,
	"array":  ExpressionBaseTypeArray,
	"map":    ExpressionBaseTypeMap,
}

func ParseExpressionBaseType(s string) (ExpressionBaseType, bool) {
	t, ok := expressionBaseTypeMap[s]
	return t, ok
}

type ExpressionType interface {
	BaseType() ExpressionBaseType
	KeyType() ExpressionBaseType
	ValueType() ExpressionBaseType
	Equals(other ExpressionType) bool
	String() string
}

type expressionType struct {
	baseType  ExpressionBaseType
	keyType   ExpressionBaseType
	valueType ExpressionBaseType
}

func NewExpressionType(baseType ExpressionBaseType, keyType ExpressionBaseType, valueType ExpressionBaseType) ExpressionType {
	return &expressionType{
		baseType:  baseType,
		keyType:   keyType,
		valueType: valueType,
	}
}

func (e *expressionType) BaseType() ExpressionBaseType {
	return e.baseType
}

func (e *expressionType) KeyType() ExpressionBaseType {
	return e.keyType
}

func (e *expressionType) ValueType() ExpressionBaseType {
	return e.valueType
}

func (e *expressionType) Equals(other ExpressionType) bool {
	return e.baseType == other.BaseType() &&
		e.keyType == other.KeyType() &&
		e.valueType == other.ValueType()
}

func (e *expressionType) String() string {
	if e.baseType == ExpressionBaseTypeArray {
		return string(e.valueType) + "[]"
	}

	if e.baseType == ExpressionBaseTypeMap {
		return "map[" + string(e.keyType) + ", " + string(e.valueType) + "]"
	}

	return string(e.baseType)
}

var _arrayTypeRegex = regexp.MustCompile(`^\s*(\w+)\[\]\s*$`)
var _mapTypeRegex = regexp.MustCompile(`^\s*map\[\s*(.*)\s*,\s*(.*)\s*\]\s*$`)

func ParseExpressionType(s string) (ExpressionType, bool) {
	s = strings.TrimSpace(s)

	if strings.HasSuffix(s, "[]") {
		matches := _arrayTypeRegex.FindStringSubmatch(s)
		if len(matches) != 2 {
			return nil, false
		}

		valueType, ok := ParseExpressionBaseType(s[:len(s)-2])
		if !ok {
			return nil, false
		}
		return NewExpressionType(ExpressionBaseTypeArray, ExpressionBaseTypeInt, valueType), true
	}

	if strings.HasPrefix(s, "map[") {
		matches := _mapTypeRegex.FindStringSubmatch(s)
		if len(matches) != 3 {
			return nil, false
		}

		keyType, ok := ParseExpressionBaseType(matches[1])
		if !ok {
			return nil, false
		}
		valueType, ok := ParseExpressionBaseType(matches[2])
		if !ok {
			return nil, false
		}
		return NewExpressionType(ExpressionBaseTypeMap, keyType, valueType), true
	}

	if t, ok := ParseExpressionBaseType(s); ok {
		return NewExpressionType(t, "", t), ok
	}

	return nil, false
}
