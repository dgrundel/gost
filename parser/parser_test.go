package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestParseFromFile(t *testing.T) {

// 	content, err := os.ReadFile("test.html")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	document, err := Parse(string(content))
// 	assert.NoError(t, err)
// 	assert.NotNil(t, document)
// 	if document != nil {
// 		assert.Equal(t, "html", document.String())
// 	}
// }

func TestParseSmall(t *testing.T) {

	html := `
	<main>
		<p class="greeting">Hello, world!</p>
		<img src="#" alt="empty">
	</main>
	`

	document, err := Parse(html)
	assert.NoError(t, err)
	assert.NotNil(t, document)
	if document != nil {
		assert.Equal(t, html, document.OuterHTML())
	}
}
