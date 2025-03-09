package main

import (
	"fmt"
	"gost/parser"
	"strings"
)

func main() {
	r := strings.NewReader("<h1>Hello, world!</h1>")
	document, err := parser.Parse(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(document.OuterHTML())
}
