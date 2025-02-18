package parser

import (
	"bytes"
	"os"
	"testing"

	"golang.org/x/net/html"
)

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
