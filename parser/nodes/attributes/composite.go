package attributes

type AttributeValueComposite interface {
	AttributeValue
}

type attributeValueComposite struct {
}

func NewAttributeValueComposite(s string) (AttributeValueComposite, error) {
	return &attributeValueComposite{}, nil
}

func (c *attributeValueComposite) OuterHTML() string {
	return ""
}

func (c *attributeValueComposite) IsEmpty() bool {
	return false
}
