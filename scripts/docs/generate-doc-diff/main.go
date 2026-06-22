package main

import (
	"fmt"
	"os"
	"strings"
)

func readFile(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return string(bytes)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/docs/generate-doc-diff <from.md> <to.md>")
		os.Exit(1)
	}
	fromLines := strings.Split(readFile(os.Args[1]), "\n")
	toLines := strings.Split(readFile(os.Args[2]), "\n")
	maxLen := len(fromLines)
	if len(toLines) > maxLen {
		maxLen = len(toLines)
	}
	for i := 0; i < maxLen; i++ {
		var fromLine, toLine string
		hasFrom := i < len(fromLines)
		hasTo := i < len(toLines)
		if hasFrom {
			fromLine = fromLines[i]
		}
		if hasTo {
			toLine = toLines[i]
		}
		switch {
		case hasFrom && hasTo && fromLine == toLine:
			continue
		case hasFrom && hasTo:
			fmt.Printf("- %s\n", fromLine)
			fmt.Printf("+ %s\n", toLine)
		case hasFrom:
			fmt.Printf("- %s\n", fromLine)
		case hasTo:
			fmt.Printf("+ %s\n", toLine)
		}
	}
}
