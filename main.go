package main

import (
	"bufio"
	"flag"
	"fmt"
	"gost/generators/typescript"
	"gost/parser"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func main() {
	flag.Parse()

	var files []string
	for _, pattern := range flag.Args() {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			panic(err)
		}
		files = append(files, matches...)
	}

	fmt.Printf("Found %d files\n", len(files))

	for _, file := range files {
		fmt.Println(file)

		reader, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		document, err := parser.Parse(bufio.NewReader(reader))
		if err != nil {
			panic(err)
		}

		outputFile := strings.TrimSuffix(file, filepath.Ext(file)) + ".ts"
		writer, err := os.Create(outputFile)
		if err != nil {
			panic(err)
		}
		defer writer.Close()

		typescript.Generate(document, writer)

		fmt.Println(outputFile)
	}
}
