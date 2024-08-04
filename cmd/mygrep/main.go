package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	//	"strings"
	"unicode/utf8"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool
	plen := utf8.RuneCountInString(pattern)

	if plen == 1 {
		ok = bytes.ContainsAny(line, pattern)
	} else if plen == 2 {
		switch pattern[1] {
		case 'w':
			ok = bytes.ContainsFunc(bytes.ToLower(line), func(r rune) bool {
				return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
			})
			break
		case 'd':
			ok = bytes.ContainsAny(line, "0123456789")
			break

		}
	} else if pattern[0] == '[' && pattern[len(pattern)-1] == ']' {
		group := pattern[1 : len(pattern)-1]
		ok = bytes.ContainsAny(line, group)
	} else {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	return ok, nil
}
