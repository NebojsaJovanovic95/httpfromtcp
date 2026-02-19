package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(line string) (RequestLine, error) {
	requestLine := strings.Split(line, "\r\n")[0]
	items := strings.Split(requestLine, " ")
	if len(items) != 3 {
		return RequestLine{}, fmt.Errorf("request line not proper")
	}

	version := strings.TrimPrefix(items[2], "HTTP/")
	if version == items[2] { // If no prefix, invalid format
		return RequestLine{}, fmt.Errorf("invalid HTTP version format")
	}

	return RequestLine{
		HttpVersion:   version,
		RequestTarget: items[1],
		Method:        items[0],
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	reqBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(string(reqBytes))
	if err != nil {
		return nil, err
	}

	var req Request
	req.RequestLine = requestLine

	return &req, nil
}
