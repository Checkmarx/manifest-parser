package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Checkmarx/manifest-parser/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <manifest file>\n", os.Args[0])
		os.Exit(1)
	}
	manifestFile := os.Args[1]

	p := parser.ParsersFactory(manifestFile)
	if p == nil {
		log.Fatalf("Unsupported manifest type: %s", manifestFile)
	}

	pkgs, err := p.Parse(manifestFile)
	if err != nil {
		log.Fatalf("Error parsing manifest file: %v", err)
	}

	data, err := json.MarshalIndent(pkgs, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal packages to JSON: %v", err)
	}
	fmt.Println(string(data))
}
