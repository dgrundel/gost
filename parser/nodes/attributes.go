package nodes

type Iter = func(yield func(key, value string) bool)

type Attributes interface {
	GetAttribute(key string) string
	SetAttribute(key, value string)
	Iterator() Iter
	All() map[string]string
}

type attrs struct {
	keys   []string
	values map[string]string
}

func NewAttributes() Attributes {
	return &attrs{
		values: make(map[string]string),
	}
}

func (a *attrs) GetAttribute(key string) string {
	return a.values[key]
}

func (a *attrs) SetAttribute(key, value string) {
	_, exists := a.values[key]
	if !exists {
		a.keys = append(a.keys, key)
	}

	a.values[key] = value
}

func (a *attrs) Iterator() Iter {
	return func(yield func(key, value string) bool) {
		for _, key := range a.keys {
			if !yield(key, a.values[key]) {
				return
			}
		}
	}
}

func (a *attrs) All() map[string]string {
	return a.values
}
