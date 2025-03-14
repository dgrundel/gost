package typescript

import (
	"fmt"
	"gost/parser/expressions"
	"gost/parser/nodes"
	"gost/parser/nodes/attributes"
	"io"
	"strings"
)

func Generate(node nodes.Node, w io.Writer) error {
	switch n := node.(type) {
	case nodes.Document:
		return generateDocument(n, w)
	case nodes.Element:
		return generateElement(n, w)
	case nodes.LoopBlock:
		return generateLoopBlock(n, w)
	case nodes.OutputBlock:
		return generateOutputBlock(n, w)
	case nodes.ConditionalBlock:
		return generateConditionalBlock(n, w)
	case nodes.TextNode:
		return generateText(n, w)
	default:
		return fmt.Errorf("unsupported node type: %T", n)
	}
}

func writeString(w io.Writer, s string) error {
	w.Write([]byte(s))
	return nil
}

func generateDocument(n nodes.Document, w io.Writer) error {
	generateHelpers(w)

	typeName := "model"
	renderFuncName := "render"
	generateType(typeName, n.GetDeclaredTypes(), w)

	var fields []string
	for name := range n.GetDeclaredTypes() {
		fields = append(fields, name)
	}

	writeString(w, "\nexport const ")
	writeString(w, renderFuncName)
	writeString(w, " = ({")
	writeString(w, strings.Join(fields, ", "))
	writeString(w, "}: ")
	writeString(w, typeName)
	writeString(w, ") => (`")

	for _, child := range n.Children() {
		Generate(child, w)
	}
	writeString(w, "`);")
	return nil
}

func generateHelpers(w io.Writer) {
	writeString(w, `const encoder = document.createElement('div');
const htmlEncode = (value: string) => {
	encoder.textContent = value;
	return encoder.innerHTML;
};
`)
}

func generateElement(n nodes.Element, w io.Writer) error {
	writeString(w, "<")
	writeString(w, n.Name())

	n.Attributes().Iterator()(func(key string, value attributes.AttributeValue) bool {
		writeString(w, " ")
		writeString(w, key)

		if value != nil && !value.IsEmpty() {
			writeString(w, "=\"${htmlEncode(`")
			generateAttributeValue(value, w)
			writeString(w, "`)}\"")
		}

		return true
	})

	spread := n.Attributes().GetSpreadAttribute()
	if spread != nil && !spread.IsEmpty() {
		writeString(w, " ")
		writeString(w, "${[...Object.entries(")
		writeString(w, spread.Key())
		writeString(w, ")].map(([k, v]) => (`${k}=\"${htmlEncode(v)}\"`)).join(' ')}")
	}

	writeString(w, ">")

	if n.IsVoid() {
		return nil
	}

	for _, child := range n.Children() {
		Generate(child, w)
	}

	writeString(w, "</")
	writeString(w, n.Name())
	writeString(w, ">")
	return nil
}

func generateText(n nodes.TextNode, w io.Writer) error {
	writeString(w, n.TextContent())
	return nil
}

func generateLoopBlock(n nodes.LoopBlock, w io.Writer) error {
	writeString(w, "${[...(Array.isArray(")
	writeString(w, n.ItemsKey())
	writeString(w, ") ? ")
	writeString(w, n.ItemsKey())
	writeString(w, ".entries() : Object.entries(")
	writeString(w, n.ItemsKey())
	writeString(w, "))].map(([")
	writeString(w, n.IndexKey())
	writeString(w, ", ")
	writeString(w, n.ValueKey())
	writeString(w, "]) => (`")

	for _, child := range n.Children() {
		Generate(child, w)
	}
	writeString(w, "`)).join('')}")
	return nil
}

func generateOutputBlock(n nodes.OutputBlock, w io.Writer) error {
	writeString(w, "${")
	writeString(w, n.Key())
	writeString(w, "}")
	return nil
}

func generateConditionalBlock(n nodes.ConditionalBlock, w io.Writer) error {
	writeString(w, "${(")
	generateBooleanExpression(n.Condition(), w)
	writeString(w, ") && (`")
	for _, child := range n.Children() {
		Generate(child, w)
	}
	next := n.Next()
	for next != nil {
		if next.Condition() != nil {
			writeString(w, "`) || (")
			generateBooleanExpression(next.Condition(), w)
			writeString(w, ") && (`")
		} else {
			writeString(w, "`) || (`")
		}
		for _, child := range next.Children() {
			Generate(child, w)
		}
		next = next.Next()
	}
	writeString(w, "`) || ''}")
	return nil
}

func generateBooleanExpression(n expressions.BooleanExpression, w io.Writer) error {
	if n.Literal() != "" {
		writeString(w, n.Literal())
		return nil
	}
	if n.Parentheses() {
		writeString(w, "(")
	}
	if n.Left() != nil {
		generateBooleanExpression(n.Left(), w)
		writeString(w, " ")
	}
	writeString(w, string(n.Operator()))
	if n.Right() != nil {
		writeString(w, " ")
		generateBooleanExpression(n.Right(), w)
	}
	if n.Parentheses() {
		writeString(w, ")")
	}
	return nil
}

func generateAttributeValue(value attributes.AttributeValue, w io.Writer) error {
	switch v := value.(type) {
	case attributes.AttributeValueString:
		writeString(w, string(v))
	case attributes.AttributeValueComposite:
		for _, value := range v.Values() {
			generateAttributeValue(value, w)
		}
	case attributes.AttributeValueExpression:
		writeString(w, "${")
		writeString(w, v.Key())
		writeString(w, "}")
	default:
		return fmt.Errorf("unsupported attribute value type: %T", v)
	}
	return nil
}

func generateType(name string, t map[string]expressions.ExpressionType, w io.Writer) error {
	writeString(w, "export interface ")
	writeString(w, name)
	writeString(w, " {\n")
	for name, t := range t {
		writeString(w, name)
		writeString(w, ": ")
		writeString(w, getTypeScriptType(t))
		writeString(w, ";\n")
	}
	writeString(w, "}")
	return nil
}

func getTypeScriptType(t expressions.ExpressionType) string {
	if t.BaseType() == expressions.ExpressionBaseTypeArray {
		return getTypeScriptBaseType(t.ValueType()) + "[]"
	}

	if t.BaseType() == expressions.ExpressionBaseTypeMap {
		return "Record<" + getTypeScriptBaseType(t.KeyType()) + "," + getTypeScriptBaseType(t.ValueType()) + ">"
	}

	return getTypeScriptBaseType(t.BaseType())
}

func getTypeScriptBaseType(t expressions.ExpressionBaseType) string {
	switch t {
	case expressions.ExpressionBaseTypeString:
		return "string"
	case expressions.ExpressionBaseTypeInt:
		return "number"
	case expressions.ExpressionBaseTypeFloat:
		return "number"
	case expressions.ExpressionBaseTypeBool:
		return "boolean"
	}
	return ""
}
