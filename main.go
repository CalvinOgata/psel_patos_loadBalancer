package main

import (
	"bufio"   // wraps stream readers and writers into a buffer
	"fmt"     // standard library for cli output
	"log"     // standard library for console output regarding concurrency and errors
	"net"     // standard library for TCP/IP connections and domain sockets
	"os"      // read files from disk
	"strings" // manipulates strings in Golang
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

	reader := bufio.NewReader(conn)             // the "raw" HTTP request
	requestLine, err := reader.ReadString('\n') // break each string after '\n'
	if err != nil {
		return
	}

	sections := strings.Fields(requestLine) // mini manual parsing of Method (POST, GET) and Path ('/')
	if len(sections) < 2 {
		return
	}
	method := sections[0] // 'GET' or 'POST' request
	path := sections[1]   // the path, in this case '/'

	if path == "/" {
		switch method {
		case "GET": // sends the HTML file across the connection
			html, err := os.ReadFile("index.html")
			if err != nil {
				sendResponse(conn, "text/plain", "500 Internal Server Error")
			}
			sendResponse(conn, "text/html", string(html))
		case "POST": // updates with the "Hello World" message, sent to JavaScript and formulated afterwards
			sendResponse(conn, "text/plain", "Hello World!")
		default:
			sendResponse(conn, "text/plain", "That's illegal...")
		}
	} else {
		sendResponse(conn, "text/plain", "404 Not Found")
	}
}

func sendResponse(conn net.Conn, contentType string, body string) {
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n\r\n"+
			"%s",
		contentType, len(body), body,
	)
	conn.Write([]byte(response))
}
