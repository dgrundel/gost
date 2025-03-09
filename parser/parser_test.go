package parser

import (
	"bytes"
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

			var buf bytes.Buffer
			document.Render(map[string]any{}, &buf)
			assert.Equal(t, expected, buf.String())
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
		model    map[string]any
	}{
		{
			name: "emit value with type",
			html: `<p>Hello, {name: string}!</p>`,
			model: map[string]any{
				"name": "John",
			},
			expected: `<p>Hello, John!</p>`,
		}, {
			name: "emit value again without redeclaration",
			html: `<p>The name is {lastName: string}, {firstName: string} {lastName}.</p>`,
			model: map[string]any{
				"firstName": "James",
				"lastName":  "Bond",
			},
			expected: `<p>The name is Bond, James Bond.</p>`,
		}, {
			name: "full attribute",
			html: `<div>
				<img src={imgSrc: string} alt={imgAlt: string}>
			</div>`,
			model: map[string]any{
				"imgSrc": "https://example.com/image.jpg",
				"imgAlt": "A photo of John",
			},
			expected: `<div>
				<img src="https://example.com/image.jpg" alt="A photo of John">
			</div>`,
		}, {
			name: "partial attribute",
			html: `<div>
				<img src={imgSrc: string} alt="A photo of {name: string}">
			</div>`,
			model: map[string]any{
				"imgSrc": "https://example.com/image.jpg",
				"name":   "John",
			},
			expected: `<div>
				<img src="https://example.com/image.jpg" alt="A photo of John">
			</div>`,
		}, {
			name: "spread attributes",
			html: `<div>
				<img {...attrs: map[string, string]}>
			</div>`,
			model: map[string]any{
				"attrs": map[string]any{
					"src": "https://example.com/image.jpg",
					"alt": "A photo of John",
				},
			},
			expected: `<div>
				<img src="https://example.com/image.jpg" alt="A photo of John">
			</div>`,
		}, {
			name: "simple if",
			html: `<div>
				{if qty > 0}You have {qty} item(s).{/if}
			</div>`,
			model: map[string]any{
				"qty": 1,
			},
			expected: `<div>
				You have 1 item(s).
			</div>`,
		}, {
			name: "if...else (1)",
			html: `<div>
				{if qty == 1}
					You have {qty} item.
				{else}
					You have {qty} items.
				{/if}
			</div>`,
			model: map[string]any{
				"qty": 1,
			},
			expected: `<div>
				You have 1 item.
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
			model: map[string]any{
				"qty": 2,
			},
			expected: `<div>
				You have 2 items.
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
			model: map[string]any{
				"qty": 1001,
			},
			expected: `<div>
				You have way too many items.
			</div>`,
		}, {
			name: "for loop",
			html: `<ul>
				{for i, item in items: string[]}
					<li data-index={i}>{item.name}</li>
				{/for}
			</ul>`,
			model: map[string]any{
				"items": []map[string]any{
					{
						"name": "Item 1",
					},
					{
						"name": "Item 2",
					},
				},
			},
			expected: `<ul>

					<li data-index="0">Item 1</li>
					<li data-index="1">Item 2</li>
				
			</ul>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.html)

			document, err := Parse(r)
			assert.NoError(t, err)
			assert.NotNil(t, document)

			var buf bytes.Buffer
			document.Render(tt.model, &buf)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
