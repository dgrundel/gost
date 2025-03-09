package nodes

type ContextData = map[string]any

type RenderContext interface {
	Get(key string) (interface{}, bool)
	WithData(data ContextData) RenderContext
}

type renderContext struct {
	data ContextData
}

func NewRenderContext(data ContextData) RenderContext {
	return &renderContext{data: data}
}

func GetTyped[T any](c RenderContext, key string) (T, bool) {
	v, ok := c.Get(key)
	if !ok {
		return *new(T), false
	}
	return v.(T), true
}

func (c *renderContext) Get(key string) (interface{}, bool) {
	v, ok := c.data[key]
	return v, ok
}

func (c *renderContext) WithData(data ContextData) RenderContext {
	newData := make(ContextData)
	for k, v := range c.data {
		newData[k] = v
	}
	for k, v := range data {
		newData[k] = v
	}
	return &renderContext{data: newData}
}
