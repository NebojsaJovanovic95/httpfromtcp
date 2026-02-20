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
	end := bytes.Index(data, []byte("\r\n\r\n"))
	if end == -1 {
		return 0, false, nil
	}

	if end == 0 {
		return 0, true, nil
	}

	line := string(data[:end])
	parts := strings.Split(line, ": ")
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("malformed header line")
	}
	if strings.Contains(parts[0], " ") || strings.Contains(parts[1], " ") {
		return 0, false, fmt.Errorf("invalid spacing header")
	}
	for _, r := range parts[0] {
		if !isValidTChar(r) {
			return 0, false, fmt.Errorf("key contains non alphanumeric runes")
		}
	}
	key := strings.ToLower(parts[0])
	value := strings.TrimSpace(parts[1])

	if _, ok := h[key]; ok {
		return 0, false, fmt.Errorf("header key already exists")
	}
	h[key] = value

	return end + 2, false, nil
}

func isValidTChar(r rune) bool {
	// ALPHA
	if unicode.IsLetter(r) {
		return true
	}
	// DIGIT
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
