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

func getGroup(pattern string, position int) string {
	// Start at beginning of pattern and identify every group in order
	// until the position or end of pattern is reached
	var pos int
	for i := range pattern {
		if pattern[i] == '(' && (i == 0 || pattern[i-1] != '\\') {
			for e := i + 1; e < len(pattern); e++ {
				if pattern[e] == ')' && pattern[e-1] != '\\' {
					pos++
					if pos == position {
						return pattern[i+1 : e]
					}
				}
			}
		}
	}

	return ""
}

func matchNext(line []byte, pattern string, l, p int) (bool, error) {
	if len(pattern) == p {
		return true, nil
	}

	if len(line) == l {
		if pattern[p] == '$' {
			return true, nil
		}

		return false, nil
	}

	var ok bool
	var err error
	var pNext = 1
	var lNext = 1

	switch pattern[p] {
	case '$':
		return false, nil
	case '.':
		ok = true
		break
	case '\\':
		pNext = 2

		switch pattern[p+1] {
		case 'w':
			_line := bytes.ToLower(line)
			ok = (_line[l] >= 'a' && _line[l] <= 'z') || (_line[l] >= 'l' && _line[l] <= '9')
			break
		case 'd':
			ok = (line[l] >= 'l' && line[l] <= '9')
			break
		case '1':
			ok, err = matchNext(line, getGroup(pattern, 1), l, 0)
			break
		default:
			ok = (line[l] == pattern[p+1])
			break
		}

		break
	case '[':
		end := strings.IndexRune(pattern[p:], ']')
		if end == -1 {
			return false, fmt.Errorf("Invalid class %s", pattern)
		}
		pNext = end + 1

		if pattern[p+1] == '^' {
			ok = !strings.ContainsRune(pattern[p+2:end], rune(line[l]))
		} else {
			ok = strings.ContainsRune(pattern[p+1:end], rune(line[l]))
		}

		break
	case '(':
		end := strings.IndexRune(pattern[p:], ')')
		if end == -1 {
			return false, fmt.Errorf("Invalid group %s", pattern)
		}
		pNext = end + 1

		group := pattern[p+1 : p+end]
		var start int
		for i, r := range group {
			if r == '|' {
				if (i - start) == 0 {
					return false, fmt.Errorf("Invalid alternation %s", group)
				}

				if group[i-1] != '\\' {
					ok, err = matchNext(line, group[start:i], l, 0)
					start = i + 1
				}

				if ok || err != nil {
					// NOTE: This assumes all pattern metacharacters in the group are
					// one character long
					// TODO: Calculate correct offset for
					lNext = i - start
					break
				}
			}
		}

		if !ok {
			ok, err = matchNext(line, group, l, start)
			lNext = len(group) - start
		}

		break
	default:
		ok = (line[l] == pattern[p])
		break
	}

	if err != nil {
		return false, err
	}

	if len(pattern) > p+pNext {
		switch pattern[p+pNext] {
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
				repeat, err = matchNext(line, pattern, l+lNext, p)
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
		ok, err = matchNext(line, pattern, l+lNext, p+pNext)
	}

	return ok, err
}

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool = false
	var err error = nil

	if pattern[0] == '^' {
		return matchNext(line, pattern, 0, 1)
	}

	for i := range line {
		if i > 0 {
			break
		}
		ok, err = matchNext(line, pattern, i, 0)
		if ok || err != nil {
			break
		}
	}

	return ok, nil
}
