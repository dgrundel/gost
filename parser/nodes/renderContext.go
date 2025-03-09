package nodes

type RenderContext interface {
	Get(key string) (interface{}, bool)
}

type renderContext struct {
	model map[string]any
}

func NewRenderContext(model map[string]any) RenderContext {
	return &renderContext{model: model}
}

func GetTyped[T any](c RenderContext, key string) (T, bool) {
	v, ok := c.Get(key)
	if !ok {
		return *new(T), false
	}
	return v.(T), true
}

func (c *renderContext) Get(key string) (interface{}, bool) {
	v, ok := c.model[key]
	return v, ok
}
