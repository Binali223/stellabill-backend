package main

import (
	"flag"
	"fmt"
	"os"

	"stellarbill-backend/internal/adr"
)

func main() {
	writeIndex := flag.Bool("write-index", false, "regenerate docs/adr/README.md")
	checkIndex := flag.Bool("check-index", true, "fail if README.md index is stale")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	if *writeIndex {
		if err := adr.WriteIndex(*root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("Wrote ADR index.")
	}

	recs, err := adr.Lint(*root, *checkIndex)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	n := 0
	for _, r := range recs {
		if !r.IsTemplate {
			n++
		}
	}
	fmt.Printf("ADR lint OK (%d decisions, template present, unique numbers).\n", n)
}
