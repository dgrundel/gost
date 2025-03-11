package nodes

type Iter = func(yield func(key string, value AttributeValue) bool)

type AttributeValue interface {
	OuterHTML() string
	IsEmpty() bool
}

type AttributeValueString string

func (s AttributeValueString) OuterHTML() string {
	return string(s)
}

func (s AttributeValueString) IsEmpty() bool {
	return s == ""
}

type AttributeValueExpression string

func (e AttributeValueExpression) OuterHTML() string {
	return "{" + string(e) + "}"
}

func (e AttributeValueExpression) IsEmpty() bool {
	return e == ""
}

type Attributes interface {
	GetAttribute(key string) AttributeValue
	SetAttribute(key string, value AttributeValue)
	Iterator() Iter
	All() map[string]AttributeValue
}

type attrs struct {
	keys   []string
	values map[string]AttributeValue
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
