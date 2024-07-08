package main

import (
	"fmt"
	"net"
)

func main() {
	addr := "localhost:3000"

	ln, err := net.Listen("tcp", addr)

	if err != nil {
		panic(err)
	}
	defer ln.Close()

	host, port, err := net.SplitHostPort(ln.Addr().String())

	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening on host: %s port: %s\n", host, port)

	for {
		conn, err := ln.Accept()

		if err != nil {
			panic(err)
		}

		go func(conn net.Conn) {
			buff := make([]byte, 1024)

			n, err := conn.Read(buff)

			if err != nil {
				fmt.Printf("Error reading: %#v\n", err)
			}

			fmt.Printf("Received message:\n%s\n", string(buff[:n]))

			conn.Write([]byte("Message received\n"))
			conn.Close()

		}(conn)

	}
}
