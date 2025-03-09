package nodes

type ExpressionBaseType string

const (
	ExpressionBaseTypeString ExpressionBaseType = "string"
	ExpressionBaseTypeInt    ExpressionBaseType = "int"
	ExpressionBaseTypeFloat  ExpressionBaseType = "float"
	ExpressionBaseTypeBool   ExpressionBaseType = "bool"
	ExpressionBaseTypeArray  ExpressionBaseType = "array"
	ExpressionBaseTypeMap    ExpressionBaseType = "map"
)

type ExpressionType struct {
	BaseType  ExpressionBaseType
	KeyType   ExpressionBaseType
	ValueType ExpressionBaseType
}

func NewExpressionType(baseType ExpressionBaseType, keyType ExpressionBaseType, valueType ExpressionBaseType) ExpressionType {
	return ExpressionType{
		BaseType:  baseType,
		KeyType:   keyType,
		ValueType: valueType,
	}
}

func (e *ExpressionType) String() string {
	if e.BaseType == ExpressionBaseTypeArray {
		return string(e.ValueType) + "[]"
	}

	if e.BaseType == ExpressionBaseTypeMap {
		return "map[" + string(e.KeyType) + ", " + string(e.ValueType) + "]"
	}

	return string(e.BaseType)
}
