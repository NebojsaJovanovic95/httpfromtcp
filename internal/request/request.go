package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type Status int

const (
	statusInitialized Status = iota
	statusParsingHeaders
	statusParsingBody
	statusDone
)

type Request struct {
	RequestLine   RequestLine
	Headers       headers.Headers
	Body          []byte
	bodyLinesRead int
	status        Status
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (rl *RequestLine) ToString() string {
	str := "Request line:\n"
	str += "- Method: " + rl.Method + "\n"
	str += "- Target: " + rl.RequestTarget + "\n"
	str += "- Version: " + rl.HttpVersion + "\n"
	return str
}

func (r *Request) ToString() string {
	str := r.RequestLine.ToString()
	str += "Headers:\n"
	str += r.Headers.ToString()
	str += "Body:\n"
	str += string(r.Body)
	return str
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, 8)
	readTo := 0

	r := &Request{
		status:  statusInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}

	for {
		if readTo == len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer[:readTo])
			buffer = newBuf
		}

		n, err := reader.Read(buffer[readTo:])
		if n > 0 {
			readTo += n
		}

		consumed, parseErr := r.parse(buffer[:readTo])
		if parseErr != nil {
			return nil, parseErr
		}

		if consumed > 0 {
			copy(buffer, buffer[consumed:readTo])
			readTo -= consumed
		}

		if r.status == statusDone {
			return r, nil
		}

		if err == io.EOF {
			if consumed == 0 {
				return nil, fmt.Errorf("incomplete request")
			}
		}
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalConsumed := 0

	for r.status != statusDone {
		n, err := r.parseSingle(data[totalConsumed:])
		if err != nil {
			return totalConsumed, err
		}

		if n == 0 {
			break
		}

		totalConsumed += n
	}

	return totalConsumed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.status {

	case statusInitialized:
		rl, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = rl
		r.status = statusParsingHeaders
		return consumed, nil

	case statusParsingHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		if done {
			r.status = statusParsingBody
		}

		return n, nil

	case statusParsingBody:
		contentLengthString, ok := r.Headers.Get("content-length")
		if !ok {
			r.status = statusDone
			return len(data), nil
		}

		contentLength, err := strconv.Atoi(contentLengthString)
		if err != nil {
			return 0, fmt.Errorf("malformed content-length: %s", err)
		}

		r.Body = append(r.Body, data...)
		r.bodyLinesRead += len(data)
		if r.bodyLinesRead > contentLength {
			return 0, fmt.Errorf("content-length is too large")
		}

		if r.bodyLinesRead == contentLength {
			r.status = statusDone
		}

		return len(data), nil

	}

	return 0, fmt.Errorf("unknown state")
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	end := bytes.Index(data, []byte("\r\n"))
	if end == -1 {
		return RequestLine{}, 0, nil
	}

	line := string(data[:end])
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("malformed request line")
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return RequestLine{}, 0, fmt.Errorf("invalid http version")
	}

	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   strings.TrimPrefix(parts[2], "HTTP/"),
	}

	return rl, end + 2, nil
}
