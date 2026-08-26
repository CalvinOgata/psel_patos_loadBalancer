package main

import (
	"fmt" // standard library for cli output
	"log" // standard library for console output regarding concurrency and errors
	"net" // standard library for TCP/IP connections and domain sockets
)

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}

	defer listener.Close()

	fmt.Println("TCP Server listening on http://localhost:8080")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	httpResponse := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"Connection: close\r\n" +
		"\r\n" +
		"Hello, World!"

	_, err := conn.Write([]byte(httpResponse))

	if err != nil {
		log.Printf("Failed to write to connection: %v", err)
	}
}
