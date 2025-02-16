package parser

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
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

func BenchmarkCustomParser(b *testing.B) {
	content, err := os.ReadFile("test.html")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(content)
		_, err = Parse(reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkBuiltInParser(b *testing.B) {
	content, err := os.ReadFile("test.html")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(content)
		_, err = html.Parse(reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}
