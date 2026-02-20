package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	end := bytes.Index(data, []byte("\r\n"))
	if end == -1 {
		return 0, false, nil
	}

	if end == 0 {
		return 2, true, nil
	}

	line := string(data[:end])

	if line != strings.TrimSpace(line) {
		return 0, false, fmt.Errorf("invalid spacing header")
	}

	colonIndex := strings.Index(line, ": ")
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("malformed header line")
	}

	key := line[:colonIndex]
	value := line[colonIndex+2:]

	if strings.HasSuffix(key, " ") {
		return 0, false, fmt.Errorf("invalid spacing header")
	}

	if key == "" {
		return 0, false, fmt.Errorf("empty header key")
	}

	for _, r := range key {
		if !isValidTChar(r) {
			return 0, false, fmt.Errorf("invalid header key")
		}
	}

	key = strings.ToLower(key)

	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + value
	} else {
		h[key] = value
	}

	return end + 2, false, nil
}

func isValidTChar(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if unicode.IsDigit(r) {
		return true
	}

	switch r {
	case '!', '#', '$', '%', '&', '\'', '*',
		'+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}

	return false
}
