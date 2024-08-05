package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	//"strings"
	//"unicode/utf8"
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

func matchNext(line []byte, pattern string) (bool, error) {
	if len(line) == 0 {
		if len(pattern) == 0 {
			return true, nil
		}

		return false, nil
	}

	if len(pattern) == 0 {
		return true, nil
	}

	var ok bool = false
	var err error = nil

	switch pattern[0] {
	case '\\':
		switch pattern[1] {
		case 'w':
			_line := bytes.ToLower(line)
			ok = (_line[0] >= 'a' && _line[0] <= 'z') || (_line[0] >= '0' && _line[0] <= '9')
			break
		case 'd':
			ok = (line[0] >= '0' && line[0] <= '9')
			break
		case '\\':
			ok = (line[0] == pattern[1])
			break
		default:
			return false, fmt.Errorf("Invalid escape sequence %s", pattern[:2])
		}

		if ok {
			ok, err = matchNext(line[1:], pattern[2:])
		}
		break
	case '[':
		end := strings.IndexRune(pattern, ']')
		if end == -1 {
			return false, fmt.Errorf("Invalid group %s", pattern)
		}

		group := pattern[1:end]
		if group[0] == '^' {
			ok = !strings.ContainsRune(group[1:], rune(line[0]))
		} else {
			ok = strings.ContainsRune(group, rune(line[0]))
		}

		if ok {
			ok, err = matchNext(line[1:], pattern[end+1:])
		}
		break
	default:
		if line[0] == pattern[0] {
			ok, err = matchNext(line[1:], pattern[1:])
		}
		break
	}

	return ok, err
}

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool = false
	var err error = nil

	for _line := line; len(_line) > 0; _line = _line[1:] {
		ok, err = matchNext(_line, pattern)
		if ok || err != nil {
			break
		}
	}

	return ok, nil
}
