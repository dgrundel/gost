package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gost/parser/nodes"
	"io"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"unicode"
)

type ParseState string

const (
	Data                       ParseState = "Data"
	TagOpen                    ParseState = "TagOpen"
	EndTagOpen                 ParseState = "EndTagOpen"
	TagName                    ParseState = "TagName"
	SelfClosingStartTag        ParseState = "SelfClosingStartTag"
	BeforeAttributeName        ParseState = "BeforeAttributeName"
	SpreadAttribute            ParseState = "SpreadAttribute"
	AttributeName              ParseState = "AttributeName"
	AfterAttributeName         ParseState = "AfterAttributeName"
	BeforeAttributeValue       ParseState = "BeforeAttributeValue"
	AttributeValueDoubleQuoted ParseState = "AttributeValueDoubleQuoted"
	AttributeValueSingleQuoted ParseState = "AttributeValueSingleQuoted"
	AttributeValueExpression   ParseState = "AttributeValueExpression"
	AttributeValueUnquoted     ParseState = "AttributeValueUnquoted"
	AfterAttributeValueQuoted  ParseState = "AfterAttributeValueQuoted"
	RawText                    ParseState = "RawText"
	RawTextLessThanSign        ParseState = "RawTextLessThanSign"
	RawTextEndTagOpen          ParseState = "RawTextEndTagOpen"
	RawTextEndTagName          ParseState = "RawTextEndTagName"
	MarkupDeclarationOpen      ParseState = "MarkupDeclarationOpen"
	Comment                    ParseState = "Comment"
	ExpressionName             ParseState = "ExpressionName"
	EndExpression              ParseState = "EndExpression"
	IfConditionalExpression    ParseState = "IfConditionalExpression"
	ElseConditionalExpression  ParseState = "ElseConditionalExpression"
	ForLoopExpression          ParseState = "ForLoopExpression"
	OutputExpressionKey        ParseState = "OutputExpressionKey"
	OutputExpressionType       ParseState = "OutputExpressionType"

	// CommentEndDash             ParseState = "CommentEndDash"
	// CommentEnd                 ParseState = "CommentEnd"
	// CommentStart               ParseState = "CommentStart"
	// CommentStartDash           ParseState = "CommentStartDash"
	// CommentEndBang             ParseState = "CommentEndBang"
	// Doctype                    ParseState = "Doctype"
	// BeforeDoctypeName          ParseState = "BeforeDoctypeName"
	// DoctypeName                ParseState = "DoctypeName"
	// AfterDoctypeName           ParseState = "AfterDoctypeName"
)

var _forLoopRegex = regexp.MustCompile(`^\s*(\w+),\s*(\w+)\s+in\s+(\w+)\s*(?:\:\s*([A-Za-z0-9_\]\[-]+))?\s*$`)

var _rawTextElements = map[string]bool{
	"script":   true,
	"style":    true,
	"textarea": true,
}

type tag struct {
	name       strings.Builder
	attrName   strings.Builder
	attrValue  strings.Builder
	attributes nodes.Attributes
	endTag     bool
}

func newTag() *tag {
	return &tag{
		attributes: nodes.NewAttributes(),
	}
}

type parseContext struct {
	Rune     rune
	Position int
	Buf      bytes.Buffer
	Temp     bytes.Buffer
	State    ParseState
	Parent   nodes.Node
	Tag      *tag
	Document nodes.Document
}

var _parseStateHandlers = map[ParseState](func(ctx *parseContext) error){
	Data:                       handleData,
	TagOpen:                    handleTagOpen,
	EndTagOpen:                 handleEndTagOpen,
	TagName:                    handleTagName,
	SelfClosingStartTag:        handleSelfClosingStartTag,
	BeforeAttributeName:        handleBeforeAttributeName,
	SpreadAttribute:            handleSpreadAttribute,
	AttributeName:              handleAttributeName,
	AfterAttributeName:         handleAfterAttributeName,
	BeforeAttributeValue:       handleBeforeAttributeValue,
	AttributeValueDoubleQuoted: handleAttributeValueDoubleQuoted,
	AttributeValueSingleQuoted: handleAttributeValueSingleQuoted,
	AttributeValueExpression:   handleAttributeValueExpression,
	AttributeValueUnquoted:     handleAttributeValueUnquoted,
	AfterAttributeValueQuoted:  handleAfterAttributeValueQuoted,
	RawText:                    handleRawText,
	RawTextLessThanSign:        handleRawTextLessThanSign,
	RawTextEndTagOpen:          handleRawTextEndTagOpen,
	RawTextEndTagName:          handleRawTextEndTagName,
	MarkupDeclarationOpen:      handleMarkupDeclarationOpen,
	Comment:                    handleComment,
	ExpressionName:             handleExpressionName,
	EndExpression:              handleEndExpression,
	IfConditionalExpression:    handleIfConditionalExpression,
	ElseConditionalExpression:  handleElseConditionalExpression,
	ForLoopExpression:          handleForLoopExpression,
	OutputExpressionKey:        handleOutputExpressionKey,
	OutputExpressionType:       handleOutputExpressionType,
}

func handleData(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '<':
		if ctx.Buf.Len() > 0 {
			text := nodes.NewTextNode(ctx.Buf.String())
			ctx.Parent.Append(text)
			ctx.Buf.Reset()
		}
		ctx.State = TagOpen
	case '{':
		if ctx.Buf.Len() > 0 {
			text := nodes.NewTextNode(ctx.Buf.String())
			ctx.Parent.Append(text)
			ctx.Buf.Reset()
		}
		ctx.Temp.Reset() // reset temp buffer for expression
		ctx.State = ExpressionName
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleTagOpen(ctx *parseContext) error {
	if ctx.Tag != nil {
		return parseErr(ctx, "tag open with incomplete tag")
	}
	ctx.Tag = newTag()

	r := ctx.Rune
	switch {
	case r == '/':
		ctx.State = EndTagOpen
	case r >= 'A' && r <= 'Z':
		ctx.Tag.name.WriteRune(unicode.ToLower(r))
		ctx.State = TagName
	case r >= 'a' && r <= 'z':
		ctx.Tag.name.WriteRune(r)
		ctx.State = TagName
	case r == '!':
		ctx.Tag = nil
		ctx.Temp.Reset()
		ctx.State = MarkupDeclarationOpen
	default:
		return parseErr(ctx, "unexpected rune")
	}
	return nil
}

func handleEndTagOpen(ctx *parseContext) error {
	r := ctx.Rune
	if ctx.Tag == nil {
		return parseErr(ctx, "tag open with incomplete tag")
	}
	ctx.Tag.endTag = true
	switch {
	case r >= 'A' && r <= 'Z':
		ctx.Tag.name.WriteRune(unicode.ToLower(r))
		ctx.State = TagName
	case r >= 'a' && r <= 'z':
		ctx.Tag.name.WriteRune(r)
		ctx.State = TagName
	default:
		return parseErr(ctx, "unexpected rune")
	}
	return nil
}

func handleTagName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		ctx.State = BeforeAttributeName
	case r == '{':
		ctx.State = SpreadAttribute
	case r == '/':
		ctx.State = SelfClosingStartTag
	case r == '>':
		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	case r >= 'A' && r <= 'Z':
		ctx.Tag.name.WriteRune(unicode.ToLower(r))
	default:
		ctx.Tag.name.WriteRune(r)
	}
	return nil
}

func handleSelfClosingStartTag(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '>':
		_, err := applyTag(ctx, true)
		if err != nil {
			return err
		}
		ctx.State = Data
	default:
		return parseErr(ctx, "unexpected rune")
	}
	return nil
}

func handleBeforeAttributeName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '{':
		ctx.State = SpreadAttribute
	case r == '/':
		ctx.State = SelfClosingStartTag
	case r == '>':
		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	case r >= 'A' && r <= 'Z':
		ctx.Tag.attrName.WriteRune(unicode.ToLower(r))
		ctx.State = AttributeName
	default:
		ctx.Tag.attrName.WriteRune(r)
		ctx.State = AttributeName
	}
	return nil
}

func handleSpreadAttribute(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case r == '}':
		ctx.Tag.attributes.SetSpreadAttribute(nodes.AttributeValueSpread(ctx.Buf.String()))
		ctx.Buf.Reset()
		ctx.State = AfterAttributeValueQuoted
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleAttributeName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		ctx.State = AfterAttributeName
	case r == '/':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}
		ctx.State = SelfClosingStartTag
	case r == '=':
		ctx.State = BeforeAttributeValue
	case r == '>':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	case r >= 'A' && r <= 'Z':
		ctx.Tag.attrName.WriteRune(unicode.ToLower(r))
	default:
		ctx.Tag.attrName.WriteRune(r)
	}
	return nil
}

func handleAfterAttributeName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '/':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}
		ctx.State = SelfClosingStartTag
	case r == '=':
		ctx.State = BeforeAttributeValue
	case r == '>':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	case r >= 'A' && r <= 'Z': // start a new attribute
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		ctx.Tag.attrName.WriteRune(unicode.ToLower(r))
		ctx.State = AttributeName
	default: // start a new attribute
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		ctx.Tag.attrName.WriteRune(r)
		ctx.State = AttributeName
	}
	return nil
}

func handleBeforeAttributeValue(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '"':
		ctx.State = AttributeValueDoubleQuoted
	case r == '\'':
		ctx.State = AttributeValueSingleQuoted
	case r == '{':
		ctx.State = AttributeValueExpression
	case r == '>':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	default:
		ctx.Tag.attrValue.WriteRune(r)
		ctx.State = AttributeValueUnquoted
	}
	return nil
}

func handleAttributeValueDoubleQuoted(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '"':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}
		ctx.State = AfterAttributeValueQuoted
	default:
		ctx.Tag.attrValue.WriteRune(r)
	}
	return nil
}

func handleAttributeValueSingleQuoted(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '\'':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}
		ctx.State = AfterAttributeValueQuoted
	default:
		ctx.Tag.attrValue.WriteRune(r)
	}
	return nil
}

func handleAttributeValueExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '}':
		err := applyAttr(ctx, true)
		if err != nil {
			return err
		}
		ctx.State = AfterAttributeValueQuoted
	default:
		ctx.Tag.attrValue.WriteRune(r)
	}
	return nil
}

func handleAttributeValueUnquoted(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}
		ctx.State = BeforeAttributeName
	case r == '>':
		err := applyAttr(ctx, false)
		if err != nil {
			return err
		}

		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	default:
		ctx.Tag.attrValue.WriteRune(r)
	}
	return nil
}

func handleAfterAttributeValueQuoted(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		ctx.State = BeforeAttributeName
	case r == '/':
		ctx.State = SelfClosingStartTag
	case r == '>':
		elem, err := applyTag(ctx, false)
		if err != nil {
			return err
		}

		if elem != nil && _rawTextElements[elem.Name()] {
			ctx.State = RawText
		} else {
			ctx.State = Data
		}
	default:
		return parseErr(ctx, "unexpected rune")
	}
	return nil
}

func handleRawText(ctx *parseContext) error {
	switch ctx.Rune {
	case '<':
		ctx.State = RawTextLessThanSign
	default:
		ctx.Buf.WriteRune(ctx.Rune)
		ctx.State = RawText
	}
	return nil
}

func handleRawTextLessThanSign(ctx *parseContext) error {
	switch ctx.Rune {
	case '/':
		ctx.Temp.Reset()
		ctx.State = RawTextEndTagOpen
	default:
		ctx.Buf.WriteByte('<')
		return handleRawText(ctx)
	}
	return nil
}

func handleRawTextEndTagOpen(ctx *parseContext) error {
	switch {
	case ctx.Rune >= 'A' && ctx.Rune <= 'Z':
		if ctx.Tag != nil {
			return parseErr(ctx, "tag open with incomplete tag")
		}
		ctx.Tag = newTag()
		ctx.Tag.endTag = true
		ctx.Tag.name.WriteRune(unicode.ToLower(ctx.Rune))
		ctx.Temp.WriteRune(ctx.Rune)
		ctx.State = RawTextEndTagName
	case ctx.Rune >= 'a' && ctx.Rune <= 'z':
		if ctx.Tag != nil {
			return parseErr(ctx, "tag open with incomplete tag")
		}
		ctx.Tag = newTag()
		ctx.Tag.endTag = true
		ctx.Tag.name.WriteRune(ctx.Rune)
		ctx.Temp.WriteRune(ctx.Rune)
		ctx.State = RawTextEndTagName
	default:
		ctx.Buf.WriteString("</")
		return handleRawText(ctx)
	}
	return nil
}

func handleRawTextEndTagName(ctx *parseContext) error {
	switch {
	case unicode.IsSpace(ctx.Rune):
		if ctx.Tag == nil {
			return parseErr(ctx, "end tag name with nil tag")
		}
		if ctx.Tag.name.String() == ctx.Parent.Name() {
			ctx.State = BeforeAttributeName
		} else {
			ctx.Buf.WriteString("</")
			ctx.Buf.Write(ctx.Temp.Bytes())
			return handleRawText(ctx)
		}
	case ctx.Rune == '/':
		if ctx.Tag == nil {
			return parseErr(ctx, "end tag name with nil tag")
		}
		if ctx.Tag.name.String() == ctx.Parent.Name() {
			ctx.State = SelfClosingStartTag
		} else {
			ctx.Buf.WriteString("</")
			ctx.Buf.Write(ctx.Temp.Bytes())
			return handleRawText(ctx)
		}
	case ctx.Rune == '>':
		if ctx.Tag == nil {
			return parseErr(ctx, "end tag name with nil tag")
		}
		if ctx.Tag.name.String() == ctx.Parent.Name() {
			if ctx.Buf.Len() > 0 {
				text := nodes.NewTextNode(ctx.Buf.String())
				ctx.Parent.Append(text)
			}
			ctx.Buf.Reset()
			_, err := applyTag(ctx, false)
			if err != nil {
				return err
			}
			ctx.State = Data
		} else {
			ctx.Buf.WriteString("</")
			ctx.Buf.Write(ctx.Temp.Bytes())
			return handleRawText(ctx)
		}
	case ctx.Rune >= 'A' && ctx.Rune <= 'Z':
		if ctx.Tag == nil {
			return parseErr(ctx, "end tag name with nil tag")
		}
		ctx.Tag.name.WriteRune(unicode.ToLower(ctx.Rune))
		ctx.Temp.WriteRune(ctx.Rune)
	case ctx.Rune >= 'a' && ctx.Rune <= 'z':
		if ctx.Tag == nil {
			return parseErr(ctx, "end tag name with nil tag")
		}
		ctx.Tag.name.WriteRune(unicode.ToLower(ctx.Rune))
		ctx.Temp.WriteRune(ctx.Rune)
	default:
		ctx.Buf.WriteString("</")
		ctx.Buf.Write(ctx.Temp.Bytes())
		return handleRawText(ctx)
	}
	return nil
}

func handleMarkupDeclarationOpen(ctx *parseContext) error {
	ctx.Temp.WriteRune(ctx.Rune)
	temp := ctx.Temp.String()
	switch {
	case temp == "--":
		ctx.Temp.Reset()
		ctx.State = Comment
	case len(temp) == 7 && strings.EqualFold(temp, "doctype"):
		// hack -- treat this like any other tag
		if ctx.Tag != nil {
			return parseErr(ctx, "tag open with incomplete tag")
		}
		ctx.Tag = newTag()
		ctx.Tag.name.WriteByte('!')
		ctx.Tag.name.WriteString(strings.ToLower(temp))
		ctx.State = BeforeAttributeName
	case len(temp) >= 7: // len("doctype")
		return parseErr(ctx, "invalid markup declaration")
	}
	return nil
}

func handleComment(ctx *parseContext) error {
	temp := ctx.Temp.String()
	switch {
	case temp == "--" && ctx.Rune == '>':
		if ctx.Buf.Len() > 0 {
			comment := nodes.NewComment(ctx.Buf.String())
			ctx.Parent.Append(comment)
			ctx.Buf.Reset()
		}
		ctx.Temp.Reset()
		ctx.State = Data
	case ctx.Rune == '-':
		ctx.Temp.WriteRune(ctx.Rune)
	default:
		ctx.Buf.WriteString(temp)
		ctx.Temp.Reset()
		ctx.Buf.WriteRune(ctx.Rune)
	}
	return nil
}

func handleExpressionName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		if ctx.Buf.Len() == 0 {
			break
		}
		str := ctx.Buf.String()
		if str == "if" {
			ctx.State = IfConditionalExpression
			ctx.Buf.Reset()
			break
		}
		if str == "else" {
			ctx.State = ElseConditionalExpression
			ctx.Buf.Reset()
			ctx.Temp.Reset()
			break
		}
		if str == "for" {
			ctx.State = ForLoopExpression
			ctx.Buf.Reset()
			break
		}
		// do not reset buf, it contains the variable/output key name
		ctx.State = OutputExpressionKey
	case r == ':':
		// put buffer content into temp buffer
		// temp will contain the variable/output key name
		ctx.Temp.Write(ctx.Buf.Bytes())
		ctx.Buf.Reset()
		ctx.State = OutputExpressionType
	case r == '/':
		if ctx.Buf.Len() > 0 {
			return parseErr(ctx, "invalid expression name: "+ctx.Buf.String()+"/")
		}
		ctx.State = EndExpression
	case r == '}':
		str := ctx.Buf.String()
		// if and for expressions must have content
		if str == "if" || str == "for" {
			return parseErr(ctx, "invalid empty if/for expression: "+str)
		}
		if str == "else" {
			ifexpr, ok := ctx.Parent.(nodes.ConditionalExpression)
			if !ok {
				return parseErr(ctx, "mismatched else expression")
			}
			expr := nodes.NewConditionalExpression()
			ifexpr.SetNext(expr)
			ctx.Parent = expr
		} else {
			expr := nodes.NewOutputExpression(str, "")
			ctx.Parent.Append(expr)
		}
		ctx.Buf.Reset()
		ctx.State = Data
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleEndExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '}':
		str := ctx.Buf.String()
		ctx.Buf.Reset()
		// can only close if/for expressions
		if str == "if" {
			_, ok := ctx.Parent.(nodes.ConditionalExpression)
			if !ok {
				return parseErr(ctx, "mismatched end expression: "+str)
			}
			ctx.Parent = ctx.Parent.Parent()
			ctx.State = Data
			break
		}
		if str == "for" {
			_, ok := ctx.Parent.(nodes.LoopExpression)
			if !ok {
				return parseErr(ctx, "mismatched end expression: "+str)
			}
			ctx.Parent = ctx.Parent.Parent()
			ctx.State = Data
			break
		}
		return parseErr(ctx, "invalid end expression: "+str)
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleIfConditionalExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case r == '}':
		condition := ctx.Buf.String()
		expr := nodes.NewConditionalExpression()
		expr.SetCondition(condition)
		ctx.Parent.Append(expr)
		ctx.Parent = expr
		ctx.State = Data
		ctx.Buf.Reset()
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleElseConditionalExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		if ctx.Buf.Len() > 0 {
			ctx.Buf.WriteRune(r)
		}
	case r == '}':
		if ctx.Temp.String() != "if" {
			return parseErr(ctx, "invalid else expression: "+ctx.Temp.String())
		}
		ifexpr, ok := ctx.Parent.(nodes.ConditionalExpression)
		if !ok {
			return parseErr(ctx, "mismatched else expression")
		}
		expr := nodes.NewConditionalExpression()
		ifexpr.SetNext(expr)
		condition := ctx.Buf.String()
		expr.SetCondition(condition)
		ctx.Parent = expr
		ctx.Buf.Reset()
		ctx.State = Data
	default:
		if ctx.Temp.Len() < 2 {
			ctx.Temp.WriteRune(r)
		} else {
			ctx.Buf.WriteRune(r)
		}
	}
	return nil
}

func handleForLoopExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case r == '}':
		content := ctx.Buf.String()
		matches := _forLoopRegex.FindStringSubmatch(content)
		if len(matches) != 5 {
			return parseErr(ctx, "invalid for loop expression: "+content)
		}

		// i, item in items
		indexKey := matches[1]
		itemKey := matches[2]
		collectionKey := matches[3]
		typ := matches[4]
		expr := nodes.NewLoopExpression(indexKey, itemKey, collectionKey, typ)

		ctx.Parent.Append(expr)
		ctx.Parent = expr
		ctx.Buf.Reset()
		ctx.State = Data
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleOutputExpressionKey(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == ':':
		// put buffer content into temp buffer
		// temp will contain the variable/output key name
		ctx.Temp.Write(ctx.Buf.Bytes())
		ctx.Buf.Reset()
		ctx.State = OutputExpressionType
	case r == '}':
		key := ctx.Temp.String()
		expr := nodes.NewOutputExpression(key, "")
		ctx.Parent.Append(expr)
		ctx.Buf.Reset()
		ctx.State = Data
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleOutputExpressionType(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '}':
		if ctx.Buf.Len() == 0 {
			return parseErr(ctx, "invalid output expression type: empty")
		}

		key := ctx.Temp.String()
		typ := ctx.Buf.String()
		expr := nodes.NewOutputExpression(key, typ)
		ctx.Parent.Append(expr)
		ctx.Temp.Reset()
		ctx.Buf.Reset()
		ctx.State = Data
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func debugInfo(ctx *parseContext) string {
	info := map[string]string{
		"rune":     string(ctx.Rune),
		"position": strconv.Itoa(ctx.Position),
		"buf":      ctx.Buf.String(),
		"temp":     ctx.Temp.String(),
		"state":    string(ctx.State),
		"parent":   "nil",
	}
	if ctx.Parent != nil {
		info["parent"] = ctx.Parent.Name()
	}
	if ctx.Tag != nil {
		t := ctx.Tag
		info["tag.name"] = t.name.String()
		info["tag.endTag"] = strconv.FormatBool(t.endTag)
		info["tag.attrName"] = t.attrName.String()
		info["tag.attrValue"] = t.attrValue.String()
		info["tag.attributes"] = strconv.Itoa(len(t.attributes.All()))
	} else {
		info["tag"] = "nil"
	}

	s, _ := json.Marshal(info)
	return string(s)
}

func parseErr(ctx *parseContext, message string) error {
	info := debugInfo(ctx)

	return fmt.Errorf("parse error: %s\n%s", message, info)
}

func applyTag(ctx *parseContext, void bool) (nodes.Element, error) {
	var elem nodes.Element = nil
	name := ctx.Tag.name.String()
	if name == "" {
		return nil, parseErr(ctx, "empty tag name")
	}

	if ctx.Tag.endTag {
		if name != ctx.Parent.Name() {
			return nil, parseErr(ctx, "tag mismatch")
		}
		ctx.Parent = ctx.Parent.Parent()

	} else {
		elem = nodes.NewElement(name, void)

		ctx.Tag.attributes.Iterator()(func(key string, value nodes.AttributeValue) bool {
			elem.Attributes().SetAttribute(key, value)
			return true
		})

		spread := ctx.Tag.attributes.GetSpreadAttribute()
		if !spread.IsEmpty() {
			elem.Attributes().SetSpreadAttribute(spread)
		}

		ctx.Parent.Append(elem)

		if !elem.IsVoid() {
			ctx.Parent = elem
		}
	}

	ctx.Tag = nil
	return elem, nil
}

func applyAttr(ctx *parseContext, isExpression bool) error {
	t := ctx.Tag
	if t == nil {
		return parseErr(ctx, "no tag for attribute")
	}
	name := t.attrName.String()
	value := t.attrValue.String()
	if name == "" {
		return parseErr(ctx, "empty attr name")
	}
	if isExpression {
		t.attributes.SetAttribute(name, nodes.AttributeValueExpression(value))
	} else {
		t.attributes.SetAttribute(name, nodes.AttributeValueString(value))
	}
	t.attrName.Reset()
	t.attrValue.Reset()
	return nil
}

func Parse(reader io.RuneReader) (nodes.Document, error) {
	document := nodes.NewDocument()

	ctx := &parseContext{
		State:    Data,
		Parent:   document,
		Tag:      nil,
		Document: document,
	}

	err := func() (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = parseErr(ctx, fmt.Sprintf("panic: %v\n%s", r, debug.Stack()))
			}
		}()

		for i := 0; true; i++ {
			r, _, err := reader.ReadRune()
			if err == io.EOF {
				break // End of input
			}
			if err != nil {
				return err
			}

			ctx.Rune = r
			ctx.Position = i
			// fmt.Println(debugInfo(ctx))

			err = _parseStateHandlers[ctx.State](ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}()

	if err != nil {
		return nil, err
	}

	if ctx.Buf.Len() > 0 {
		text := nodes.NewTextNode(ctx.Buf.String())
		ctx.Parent.Append(text)
		ctx.Buf.Reset()
	}

	return document, nil
}
