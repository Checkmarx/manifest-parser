package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Checkmarx/manifest-parser/pkg/parser"
)

func main() {
	versionSpecs := flag.Bool("version-specs", false, "report the declared version range (VersionSpec) of a dependency with no exact version")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("Usage: %s [-version-specs] <manifest file>\n", os.Args[0])
		os.Exit(1)
	}
	manifestFile := flag.Arg(0)

	var opts []parser.Option
	if *versionSpecs {
		opts = append(opts, parser.WithVersionSpecs())
	}

	p := parser.ParsersFactoryWithOptions(manifestFile, opts...)
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
