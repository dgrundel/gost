package main

import (
	"fmt"
	"gost/parser"
)

func main() {
	parsed, err := parser.Parse("Hello, world!")
	if err != nil {
		panic(err)
	}

	fmt.Println(parsed.Name(), parsed.TextContent())
}
