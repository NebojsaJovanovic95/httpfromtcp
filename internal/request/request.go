package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Status int

const (
	initialized Status = iota
	done
)

type Request struct {
	RequestLine RequestLine
	status      Status
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, 8)
	readtoIndex := 0

	r := &Request{
		status: initialized,
	}

	for r.status != done {

		if readtoIndex == len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		n, err := reader.Read(buffer[readtoIndex:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		readtoIndex += n

		consumed, err := r.parse(buffer[:readtoIndex])
		if err != nil {
			return nil, err
		}

		if consumed > 0 {
			copy(buffer, buffer[consumed:readtoIndex])
			readtoIndex -= consumed
		}
	}

	if r.status != done {
		return nil, fmt.Errorf("incomplete request")
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.status {
	case initialized:
		rl, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = rl
		r.status = done
		return consumed, nil
	case done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	end := bytes.Index(data, []byte("\r\n"))
	if end == -1 {
		// Need more data
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
