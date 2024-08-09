package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type grepState struct {
	groups []string
}

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

func (g *grepState) matchNext(line []byte, pattern string, l, p int) (int, error) {
	var err error
	var pNext = 1
	var lNext = 0

	//fmt.Printf("%s | %s\n", line[l:], pattern[p:])

	switch pattern[p] {
	case '$':
		break
	case '.':
		lNext = 1
		break
	case '\\':
		pNext = 2

		switch pattern[p+1] {
		case 'w':
			_line := bytes.ToLower(line)
			if (_line[l] >= 'a' && _line[l] <= 'z') || (_line[l] >= '0' && _line[l] <= '9') {
				lNext = 1
			}
			break
		case 'd':
			if line[l] >= '0' && line[l] <= '9' {
				lNext = 1
			}
			break
		default:
			if pattern[p+1] > '0' && pattern[p+1] <= '9' {
				var end int
				for end = p + 1; end <= len(pattern); end++ {
					if !(end >= '0' && end <= '9') {
						break
					}
				}

				var gi int
				if p+1 == end {
					gi, _ = strconv.Atoi(string(pattern[end]))
				} else {
					gi, _ = strconv.Atoi(string(pattern[p+1 : end]))
				}
				gi--

				if gi >= 0 && gi < len(g.groups) {
					lNext, err = g.matchNext(line, g.groups[gi], l, 0)
				} else {
					return 0, fmt.Errorf("Invalid group reference \\%d", pattern[p+1:])
				}
			} else if line[l] == pattern[p+1] {
				lNext = 1
			}

			break
		}

		break
	case '[':
		end := strings.IndexRune(pattern[p:], ']')
		if end == -1 {
			return 0, fmt.Errorf("Invalid class %s", pattern)
		}
		pNext = end + 1

		if pattern[p+1] == '^' {
			if !strings.ContainsRune(pattern[p+2:p+end], rune(line[l])) {
				lNext = 1
			}
		} else if strings.ContainsRune(pattern[p+1:p+end], rune(line[l])) {
			lNext = 1
		}

		break
	case '(':
		end := strings.IndexRune(pattern[p:], ')')
		if end == -1 {
			return 0, fmt.Errorf("Invalid group %s", pattern)
		}
		pNext = end + 1

		group := pattern[p+1 : p+end]
		var start int
		for i, r := range group {
			if r == '|' {
				if (i - start) == 0 {
					return 0, fmt.Errorf("Invalid alternation %s", group)
				}

				if group[i-1] != '\\' {
					lNext, err = g.matchNext(line, group[start:i], l, 0)
					start = i + 1
				}

				if lNext > 0 || err != nil {
					break
				}
			}
		}

		if lNext == 0 && err == nil {
			lNext, err = g.matchNext(line, group, l, start)
		}

		if lNext > 0 {
			g.groups = append(g.groups, string(line[l:l+lNext]))
		}

		break
	default:
		if line[l] == pattern[p] {
			lNext = 1
		}
		break
	}

	if err != nil {
		return 0, err
	}

	var optional bool
	if len(pattern) > p+pNext {
		switch pattern[p+pNext] {
		case '?':
			optional = true
			pNext++
			break
		case '+':
			if lNext == 0 {
				break
			}

			var repeat int
			for l+lNext < len(line) {
				repeat, err = g.matchNext(line, pattern[p:p+pNext], l+lNext, 0)
				if repeat > 0 && err == nil {
					lNext += repeat
				} else {
					break
				}
			}

			pNext++
			break
		}
	}

	if lNext > 0 || optional {
		if len(pattern) == p+pNext {
			return lNext, nil
		}

		if len(line) == l+lNext {
			if pattern[p+pNext] == '$' {
				return lNext, nil
			}

			return 0, nil
		}

		matched, err := g.matchNext(line, pattern, l+lNext, p+pNext)
		if matched > 0 {
			return lNext + matched, err
		}
	}

	return 0, nil
}

func matchLine(line []byte, pattern string) (bool, error) {
	var matched int
	var err error

	if pattern[0] == '^' {
		var grep grepState
		matched, err = grep.matchNext(line, pattern, 0, 1)
		return matched > 0, err
	}

	for i := range line {
		var grep grepState
		matched, err = grep.matchNext(line, pattern, i, 0)
		if matched > 0 || err != nil {
			break
		}
	}

	return matched > 0, err
}
