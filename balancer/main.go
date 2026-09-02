package main

import (
	"bufio"   // wraps stream readers and writers into a buffer
	"fmt"     // standard library for cli output
	"io"      // copies bytes straight from one connection into another
	"log"     // standard library for console output regarding concurrency and errors
	"net"     // standard library for TCP/IP connections and domain sockets
	"os"      // read files from disk
	"strings" // manipulates strings in Golang
)

const backend = "localhost:8081" // the backend server that actually holds the images

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to start TCP listener: %v", err)
	}
	defer listener.Close()

	fmt.Println("TCP Server listening on http://localhost:8080")

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	reader := bufio.NewReader(connection)       // the "raw" HTTP request
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

	switch {
	case path == "/":
		serveIndex(connection, method)
	case path == "/random-image", strings.HasPrefix(path, "/images/"):
		proxyRequest(connection, method, path) // handed over to the backend server
	default:
		sendResponse(connection, "404 Not Found", "text/plain", "404 Not Found")
	}
}

func serveIndex(connection net.Conn, method string) {
	switch method {
	case "GET": // sends the HTML file across the connection
		html, err := os.ReadFile("index.html")
		if err != nil {
			sendResponse(connection, "500 Internal Server Error", "text/plain", "Could not read index.html")
			return
		}
		sendResponse(connection, "200 OK", "text/html", string(html))
	case "POST": // updates with the "Hello World" message, sent to JavaScript and formulated afterwards
		sendResponse(connection, "200 OK", "text/plain", "Hello World!")
	default:
		sendResponse(connection, "405 Method Not Allowed", "text/plain", "That's illegal...")
	}
}

// forwards the request to the backend and copies its answer back, untouched
func proxyRequest(client net.Conn, method, path string) {
	backendConn, err := net.Dial("tcp", backend)
	if err != nil {
		log.Printf("Backend %s is down: %v", backend, err)
		sendResponse(client, "502 Bad Gateway", "text/plain", "No backend available")
		return
	}
	defer backendConn.Close()

	// rebuilds the request line by hand, the same way the backend parses it
	_, err = fmt.Fprintf(backendConn,
		"%s %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Connection: close\r\n\r\n",
		method, path, backend,
	)
	if err != nil {
		sendResponse(client, "502 Bad Gateway", "text/plain", "Failed to reach the backend")
		return
	}

	// the backend answers with a full HTTP response, so it goes back verbatim
	if _, err := io.Copy(client, backendConn); err != nil {
		log.Printf("Failed to relay the backend answer: %v", err)
	}
}

func sendResponse(connection net.Conn, status, contentType, body string) {
	response := fmt.Sprintf(
		"HTTP/1.1 %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n\r\n"+
			"%s",
		status, contentType, len(body), body,
	)
	connection.Write([]byte(response))
}
