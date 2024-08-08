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

	print(string(line))
}

func matchNext(line []byte, pattern string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}

	if len(line) == 0 {
		if pattern[0] == '$' {
			return true, nil
		}

		return false, nil
	}

	var ok bool
	var err error
	var pNext = 1
	var lNext = 1

	switch pattern[0] {
	case '$':
		return false, nil
	case '.':
		ok = true
		break
	case '\\':
		pNext = 2

		switch pattern[1] {
		case 'w':
			_line := bytes.ToLower(line)
			ok = (_line[0] >= 'a' && _line[0] <= 'z') || (_line[0] >= '0' && _line[0] <= '9')
			break
		case 'd':
			ok = (line[0] >= '0' && line[0] <= '9')
			break
		default:
			ok = (line[0] == pattern[1])
			break
		}

		break
	case '[':
		end := strings.IndexRune(pattern, ']')
		if end == -1 {
			return false, fmt.Errorf("Invalid class %s", pattern)
		}
		pNext = end + 1

		if pattern[1] == '^' {
			ok = !strings.ContainsRune(pattern[2:end], rune(line[0]))
		} else {
			ok = strings.ContainsRune(pattern[1:end], rune(line[0]))
		}

		break
	case '(':
		end := strings.IndexRune(pattern, ')')
		if end == -1 {
			return false, fmt.Errorf("Invalid group %s", pattern)
		}
		pNext = end + 1

		group := pattern[1:end]
		var start int
		for i, r := range group {
			if r == '|' {
				if (i - start) == 0 {
					return false, fmt.Errorf("Invalid alternation %s", group)
				}

				if group[i-1] != '\\' {
					ok, err = matchNext(line, group[start:i])
					start = i + 1
				}

				if ok || err != nil {
					break
				}
			}
		}

		if !ok {
			ok, err = matchNext(line, group[start:])
		}

		break
	default:
		ok = (line[0] == pattern[0])
		break
	}

	if err != nil {
		return false, err
	}

	if len(pattern) > pNext {
		switch pattern[pNext] {
		case '?':
			if !ok {
				ok = true
				lNext = 0
			}
			pNext++
			break
		case '+':
			if !ok {
				break
			}

			var repeat bool
			for {
				repeat, err = matchNext(line[lNext:], pattern)
				if repeat == true && err == nil {
					lNext++
				} else {
					break
				}
			}

			pNext++
			break
		}
	}

	if ok {
		ok, err = matchNext(line[lNext:], pattern[pNext:])
	}

	return ok, err
}

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool = false
	var err error = nil

	if pattern[0] == '^' {
		return matchNext(line, pattern[1:])
	}

	for _line := line; len(_line) > 0; _line = _line[1:] {
		ok, err = matchNext(_line, pattern)
		if ok || err != nil {
			break
		}
	}

	return ok, nil
}
