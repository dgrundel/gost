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
	AttributeName              ParseState = "AttributeName"
	AfterAttributeName         ParseState = "AfterAttributeName"
	BeforeAttributeValue       ParseState = "BeforeAttributeValue"
	AttributeValueDoubleQuoted ParseState = "AttributeValueDoubleQuoted"
	AttributeValueSingleQuoted ParseState = "AttributeValueSingleQuoted"
	AttributeValueUnquoted     ParseState = "AttributeValueUnquoted"
	AfterAttributeValueQuoted  ParseState = "AfterAttributeValueQuoted"
	RawText                    ParseState = "RawText"
	RawTextLessThanSign        ParseState = "RawTextLessThanSign"
	RawTextEndTagOpen          ParseState = "RawTextEndTagOpen"
	RawTextEndTagName          ParseState = "RawTextEndTagName"
	MarkupDeclarationOpen      ParseState = "MarkupDeclarationOpen"
	Comment                    ParseState = "Comment"
	ExpressionStart            ParseState = "ExpressionStart"
	EndExpression              ParseState = "EndExpression"
	ConditionalExpression      ParseState = "ConditionalExpression"
	LoopExpression             ParseState = "LoopExpression"

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

// i, item in items:type
var _loopExpressionRegex = regexp.MustCompile(`^\s*(\w+),\s*(\w+)\s+in\s+(\w+)\s*(?:\:\s*([A-Za-z0-9_\]\[-]+))?\s*$`)

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
	AttributeName:              handleAttributeName,
	AfterAttributeName:         handleAfterAttributeName,
	BeforeAttributeValue:       handleBeforeAttributeValue,
	AttributeValueDoubleQuoted: handleAttributeValueDoubleQuoted,
	AttributeValueSingleQuoted: handleAttributeValueSingleQuoted,
	AttributeValueUnquoted:     handleAttributeValueUnquoted,
	AfterAttributeValueQuoted:  handleAfterAttributeValueQuoted,
	RawText:                    handleRawText,
	RawTextLessThanSign:        handleRawTextLessThanSign,
	RawTextEndTagOpen:          handleRawTextEndTagOpen,
	RawTextEndTagName:          handleRawTextEndTagName,
	MarkupDeclarationOpen:      handleMarkupDeclarationOpen,
	Comment:                    handleComment,
	ExpressionStart:            handleExpressionStart,
	EndExpression:              handleEndExpression,
	ConditionalExpression:      handleConditionalExpression,
	LoopExpression:             handleLoopExpression,
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
		ctx.State = ExpressionStart
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

func handleAttributeName(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		ctx.State = AfterAttributeName
	case r == '/':
		err := applyAttr(ctx)
		if err != nil {
			return err
		}
		ctx.State = SelfClosingStartTag
	case r == '=':
		ctx.State = BeforeAttributeValue
	case r == '>':
		err := applyAttr(ctx)
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
		err := applyAttr(ctx)
		if err != nil {
			return err
		}
		ctx.State = SelfClosingStartTag
	case r == '=':
		ctx.State = BeforeAttributeValue
	case r == '>':
		err := applyAttr(ctx)
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
		err := applyAttr(ctx)
		if err != nil {
			return err
		}

		ctx.Tag.attrName.WriteRune(unicode.ToLower(r))
		ctx.State = AttributeName
	default: // start a new attribute
		err := applyAttr(ctx)
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
	case r == '>':
		err := applyAttr(ctx)
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
		err := applyAttr(ctx)
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
		err := applyAttr(ctx)
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
		err := applyAttr(ctx)
		if err != nil {
			return err
		}
		ctx.State = BeforeAttributeName
	case r == '>':
		err := applyAttr(ctx)
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

func handleExpressionStart(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	case r == '/':
		ctx.State = EndExpression
	default:
		ctx.Buf.WriteRune(r)
		str := ctx.Buf.String()

		if str == "if" || str == "else" {
			expr := nodes.NewConditionalExpression()
			ctx.Parent.Append(expr)
			ctx.Parent = expr
			ctx.Buf.Reset()
			ctx.State = ConditionalExpression
			break
		}

		if str == "for" {
			expr := nodes.NewLoopExpression()
			ctx.Parent.Append(expr)
			ctx.Parent = expr
			ctx.Buf.Reset()
			ctx.State = LoopExpression
			break
		}

		if len(str) >= 4 { // 4 is max length for a keyword
			return parseErr(ctx, "invalid expression keyword: "+str)
		}
	}
	return nil
}

func handleEndExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch {
	case unicode.IsSpace(r):
		break
	default:
		ctx.Buf.WriteRune(r)
		str := ctx.Buf.String()

		if str == "if" {
			if _, ok := ctx.Parent.(nodes.ConditionalExpression); !ok {
				return parseErr(ctx, "invalid expression keyword: "+str)
			}

			ctx.Parent = ctx.Parent.Parent()
			ctx.State = Data
			break
		}

		if str == "for" {
			if _, ok := ctx.Parent.(nodes.LoopExpression); !ok {
				return parseErr(ctx, "invalid expression keyword: "+str)
			}

			ctx.Parent = ctx.Parent.Parent()
			ctx.State = Data
			break
		}

		if len(str) >= 3 { // 3 is max length for end keyword (if, for)
			return parseErr(ctx, "invalid expression keyword: "+str)
		}
	}
	return nil
}

func handleConditionalExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '}':
		if ctx.Buf.Len() > 0 {
			condition := nodes.NewStringCondition(ctx.Buf.String())
			if expr, ok := ctx.Parent.(nodes.ConditionalExpression); ok {
				expr.SetCondition(condition)
			} else {
				return parseErr(ctx, "tried to set condition on non-conditional expression")
			}
			ctx.Buf.Reset()
		} else {
			return parseErr(ctx, "empty conditional expression")
		}
		ctx.State = Data
	default:
		ctx.Buf.WriteRune(r)
	}
	return nil
}

func handleLoopExpression(ctx *parseContext) error {
	r := ctx.Rune
	switch r {
	case '}':
		if ctx.Buf.Len() > 0 {
			expr, ok := ctx.Parent.(nodes.LoopExpression)
			if !ok {
				return parseErr(ctx, "tried to set fields on non-loop expression")
			}

			str := ctx.Buf.String()
			matches := _loopExpressionRegex.FindStringSubmatch(str)
			if len(matches) != 5 {
				return parseErr(ctx, "invalid loop expression: "+str)
			}

			// index, value in items:type
			expr.SetIndexKey(matches[1])
			expr.SetValueKey(matches[2])
			expr.SetItemsKey(matches[3])

			// TODO: parse type and add to document

			ctx.Buf.Reset()
		} else {
			return parseErr(ctx, "empty loop expression")
		}
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

		ctx.Tag.attributes.Iterator()(func(key, value string) bool {
			elem.SetAttribute(key, value)
			return true
		})

		ctx.Parent.Append(elem)

		if !elem.IsVoid() {
			ctx.Parent = elem
		}
	}

	ctx.Tag = nil
	return elem, nil
}

func applyAttr(ctx *parseContext) error {
	t := ctx.Tag
	if t == nil {
		return parseErr(ctx, "no tag for attribute")
	}
	name := t.attrName.String()
	value := t.attrValue.String()
	if name == "" {
		return parseErr(ctx, "empty attr name")
	}
	t.attributes.SetAttribute(name, value)
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
