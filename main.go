package main

import (
	"fmt"
	"gost/parser"
)

func main() {
	parsed, err := parser.Parse([]byte("Hello, world!"))
	if err != nil {
		panic(err)
	}

	fmt.Println(parsed.Name(), parsed.TextContent())
}
