package main

import (
	"ManifestParser/pkg/parser"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <manifest file>")
		return
	}
	manifestFile := os.Args[1]

	parser := parser.ParsersFactory(manifestFile)
	manifest, err := parser.Parse(manifestFile)

	if err != nil {
		fmt.Println("Error parsing manifest file: ", err)
		return
	}

	// print the packages as json
	fmt.Println(manifest)
}
