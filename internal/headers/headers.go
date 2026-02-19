package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	h := make(map[string]string, 8)
	return h
}
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	end := bytes.Index(data, []byte("\r\n\r\n"))
	if end == -1 {
		return 0, false, nil
	}

	line := string(data[:end])
	parts := strings.Split(line, ": ")
	if len(parts) != 2 {
		return 0, false, fmt.Errorf("malformed header line")
	}
	if strings.Contains(parts[0], " ") || strings.Contains(parts[1], " ") {
		return 0, false, fmt.Errorf("invalid spacing header")
	}

	if _, ok := h[parts[0]]; ok {
		return 0, false, fmt.Errorf("header key already exists")
	}
	h[parts[0]] = parts[1]

	return end + 2, false, nil
}
