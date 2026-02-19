package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer f.Close()
		defer close(lines)
		b := make([]byte, 8)
		var line string
		for {
			n, err := f.Read(b)
			if err != nil {
				if line != "" {
					lines <- line
					line = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				break
			}

			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- line + parts[i]
				line = ""
			}
			line += parts[len(parts)-1]
		}
	}()

	return lines
}

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("failed to start listener:", err)
		os.Exit(1)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("failed to accept connection:", err)
			continue
		}

		for line := range getLinesChannel(conn) {
			fmt.Println(line)
		}
		fmt.Println("connection closed")
	}
}
