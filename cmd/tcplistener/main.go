package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
	"os"
)

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

		rl, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("error:", err)
		}

		fmt.Println(rl.ToString())

		fmt.Println("connection closed")
	}

}
