package parser

import (
	"encoding/json"
	"fmt"
	"gost/parser/nodes"
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

var _rawTextElements = map[string]bool{
	"script": true,
	"style":  true,
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
	Buf      strings.Builder
	Temp     strings.Builder
	State    ParseState
	Parent   nodes.Node
	Tag      *tag
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
		ctx.State = AttributeValueDoubleQuoted
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
		ctx.Buf.WriteRune('<')
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
			ctx.Buf.WriteString(ctx.Temp.String())
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
			ctx.Buf.WriteString(ctx.Temp.String())
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
			ctx.Buf.WriteString(ctx.Temp.String())
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
		ctx.Buf.WriteString(ctx.Temp.String())
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
	case strings.ToLower(temp) == "doctype":
		// hack -- treat this like any other tag
		if ctx.Tag != nil {
			return parseErr(ctx, "tag open with incomplete tag")
		}
		ctx.Tag = newTag()
		ctx.Tag.name.WriteRune('!')
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

func Parse(str string) (nodes.Document, error) {

	document := nodes.NewDocument()

	ctx := &parseContext{
		State:  Data,
		Parent: document,
		Tag:    nil,
	}

	for i, r := range str {
		ctx.Rune = r
		ctx.Position = i

		// fmt.Println(debugInfo(ctx))

		err := func() (e error) {
			defer func() {
				if r := recover(); r != nil {
					e = parseErr(ctx, fmt.Sprintf("panic: %v\n%s", r, debug.Stack()))
				}
			}()

			return _parseStateHandlers[ctx.State](ctx)
		}()
		if err != nil {
			// fmt.Println(document.String())
			return nil, err
		}
	}

	if ctx.Buf.Len() > 0 {
		text := nodes.NewTextNode(ctx.Buf.String())
		ctx.Parent.Append(text)
		ctx.Buf.Reset()
	}

	return document, nil
}
