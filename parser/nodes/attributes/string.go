package attributes

type AttributeValueString string

func (s AttributeValueString) OuterHTML() string {
	return "\"" + string(s) + "\""
}

func (s AttributeValueString) IsEmpty() bool {
	return s == ""
}
