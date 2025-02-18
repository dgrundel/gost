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
