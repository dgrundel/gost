package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFullDocument(t *testing.T) {

	content, err := os.ReadFile("test.html")
	if err != nil {
		t.Fatal(err)
	}

	html := string(content)
	document, err := Parse(html)
	assert.NoError(t, err)
	assert.NotNil(t, document)
	if document != nil {
		fmt.Println(document.String())
		assert.Equal(t, html, document.OuterHTML())
	}
}
