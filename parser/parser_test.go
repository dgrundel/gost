package parser

import (
	"bytes"
	"gost/parser/expressions"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFullDocument(t *testing.T) {
	content, err := os.ReadFile("test.html")
	if err != nil {
		t.Fatal(err)
	}

	html := string(content)
	document, err := Parse(bytes.NewReader(content))
	assert.NoError(t, err)
	assert.NotNil(t, document)
	if document != nil {
		assert.Equal(t, html, document.OuterHTML())
	}
}

func TestParserTokenization(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "basic syntax",
			html: `
				<div>
					<p>Hello, world!</p>
					<ul>
						<li>item 1</li>
						<li>item 2</li>
					</ul>
				</div>`,
		}, {
			name: "void elements with slash",
			html: `
				<img src="#" alt="test" />
				Hello,<br />World!
				<hr />`,
			expected: `
				<img src="#" alt="test">
				Hello,<br>World!
				<hr>`,
		}, {
			name: "void elements without slash",
			html: `
				<img src="#" alt="test">
				Hello,<br>World!
				<hr>`,
		}, {
			name: "double quouted attrs",
			html: `<img src="#" alt="this is a test!">`,
		}, {
			name:     "single quouted attrs",
			html:     `<img src='#' alt='this is a test!'>`,
			expected: `<img src="#" alt="this is a test!">`,
		}, {
			name:     "unquouted attrs",
			html:     `<img src=# alt=test>`,
			expected: `<img src="#" alt="test">`,
		}, {
			name: "special characters",
			html: `<p>&lt;div&gt;This is a div&lt;/div&gt; 
        		<br>&amp;nbsp; &#169; &#x2022; &#x27; &#x22; &#x3C; &#x3E;</p>`,
		}, {
			name: "comments",
			html: `<p>Hello, <!--world!--></p>`,
		}, {
			name: "rawtext",
			html: `<body>
				<script type="javascript">
					for (let i = 0; i < 10; i++) {
						console.log(i);
					}
				</script>
			</body>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.html)

			document, err := Parse(r)
			assert.NoError(t, err)
			assert.NotNil(t, document)

			expected := tt.expected
			if expected == "" {
				expected = tt.html
			}

			assert.Equal(t, expected, document.OuterHTML())
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		message string
	}{
		{
			name: "unclosed tag",
			html: `<div>
				<p>Hello, wor
			</div>`,
			message: "tag mismatch",
		}, {
			name:    "invalid open tag name",
			html:    `<_>what</p>`,
			message: "unexpected rune",
		}, {
			name:    "invalid close tag name",
			html:    `<p>what</_>`,
			message: "unexpected rune",
		}, {
			name:    "invalid self-closing tag name",
			html:    `<br /_>`,
			message: "unexpected rune",
		}, {
			name:    "invalid rune after quoted attr",
			html:    `<img src="#"_>`,
			message: "unexpected rune",
		}, {
			name:    "invalid markup declaration",
			html:    `<!notadoctype>`,
			message: "invalid markup declaration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.html)

			document, err := Parse(r)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.message)
			assert.Nil(t, document)
		})
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
		types    map[string]expressions.ExpressionType
	}{
		{
			name:     "emit value with type",
			html:     `<p>Hello, {name: string}!</p>`,
			expected: `<p>Hello, {name:string}!</p>`,
			types: map[string]expressions.ExpressionType{
				"name": expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
			},
		}, {
			name:     "emit value again without redeclaration",
			html:     `<p>The name is {lastName: string}, {firstName: string} {lastName}.</p>`,
			expected: `<p>The name is {lastName:string}, {firstName:string} {lastName}.</p>`,
			types: map[string]expressions.ExpressionType{
				"lastName":  expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
				"firstName": expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
			},
		}, {
			name: "full attribute",
			html: `<div>
				<img src={imgSrc: string} alt={imgAlt: string}>
			</div>`,
			expected: `<div>
				<img src={imgSrc:string} alt={imgAlt:string}>
			</div>`,
			types: map[string]expressions.ExpressionType{
				"imgSrc": expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
				"imgAlt": expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
			},
		}, {
			name: "partial attribute",
			html: `<div>
				<img src={imgSrc: string} alt="A photo of {name: string}">
			</div>`,
			expected: `<div>
				<img src={imgSrc:string} alt="A photo of {name: string}">
			</div>`,
			types: map[string]expressions.ExpressionType{
				"imgSrc": expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
				// "name":   expressions.NewExpressionType(expressions.ExpressionBaseTypeString, "", expressions.ExpressionBaseTypeString),
			},
		}, {
			name: "spread attributes",
			html: `<div>
				<img {...attrs: map[string, string]}>
			</div>`,
			// types: map[string]expressions.ExpressionType{
			// 	"attrs": expressions.NewExpressionType(expressions.ExpressionBaseTypeMap, expressions.ExpressionBaseTypeString, expressions.ExpressionBaseTypeString),
			// },
		}, {
			name: "simple if",
			html: `<div>
				{if qty > 0}You have {qty} item(s).{/if}
			</div>`,
		}, {
			name: "simple if with type declaration",
			html: `<div>
				{if qty: int > 0}You have {qty} item(s).{/if}
			</div>`,
			expected: `<div>
				{if qty:int > 0}You have {qty} item(s).{/if}
			</div>`,
			types: map[string]expressions.ExpressionType{
				"qty": expressions.NewExpressionType(expressions.ExpressionBaseTypeInt, "", expressions.ExpressionBaseTypeInt),
			},
		}, {
			name: "if...else (1)",
			html: `<div>
				{if qty == 1}
					You have {qty} item.
				{else}
					You have {qty} items.
				{/if}
			</div>`,
		}, {
			name: "if...else (2)",
			html: `<div>
				{if qty == 1}
					You have {qty} item.
				{else}
					You have {qty} items.
				{/if}
			</div>`,
		}, {
			name: "if...else if",
			html: `<div>
				{if qty == 1}
					You have {qty} item.
				{else if qty > 1000}
					You have way too many items.
				{else}
					You have {qty} items.
				{/if}
			</div>`,
		}, {
			name: "for loop",
			html: `<ul>
				{for i, item in items: string[]}
					<li data-index={i}>{item}</li>
				{/for}
			</ul>`,
			expected: `<ul>
				{for i, item in items:string[]}
					<li data-index={i}>{item}</li>
				{/for}
			</ul>`,
			types: map[string]expressions.ExpressionType{
				"items": expressions.NewExpressionType(expressions.ExpressionBaseTypeArray, expressions.ExpressionBaseTypeInt, expressions.ExpressionBaseTypeString),
			},
		}, {
			name: "for loop without type declaration",
			html: `<ul>
				{for i, item in items}
					<li data-index={i}>{item.name}</li>
				{/for}
			</ul>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.html)

			document, err := Parse(r)
			assert.NoError(t, err)
			assert.NotNil(t, document)

			if document != nil {
				expected := tt.html
				if tt.expected != "" {
					expected = tt.expected
				}

				assert.Equal(t, expected, document.OuterHTML())

				if tt.types != nil {
					declaredTypes := document.GetDeclaredTypes()
					assert.Equal(t, len(tt.types), len(declaredTypes), "number of declared types mismatch")
					for key, expectedType := range tt.types {
						actualType, exists := declaredTypes[key]
						assert.True(t, exists, "type for %s should exist", key)
						if exists {
							assert.True(t, expectedType.Equals(actualType),
								"type mismatch for %s: expected %s, got %s",
								key, expectedType.String(), actualType.String())
						}
					}
				}

				t.Log(document.String())
			}
		})
	}
}
