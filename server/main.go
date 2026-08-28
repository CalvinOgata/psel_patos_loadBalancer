package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := flag.String("port", "8081", "Port to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Server running on http://localhost:%s\n", *port)

	for {
		connection, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	reader := bufio.NewReader(connection)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return
	}
	method, path := parts[0], parts[1]

	if method == "GET" && strings.HasPrefix(path, "/images/") {
		serveImage(connection, path)
	} else {
		sendResponse(connection, "404 Not Found", "text/plain", []byte("Image Not Found"))
		return
	}
}

func serveImage(connection net.Conn, reqPath string) {
	filename := strings.TrimPrefix(reqPath, "/images/")
	cleanPath := filepath.Join(".", "images", filepath.Clean(filename))

	imageBinary, err := os.ReadFile(cleanPath)
	if err != nil {
		sendResponse(connection, "404 Not Found", "text/plain", []byte("Image Not Found"))
		return
	}

	contentType := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(cleanPath)) {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	default:
		contentType = "error" //check on later
	}

	sendResponse(connection, "200 OK", contentType, imageBinary)
}

func sendResponse(connection net.Conn, status, contentType string, body []byte) {
	header := fmt.Sprintf(
		"HTTP/1.1 %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n\r\n",
		status, contentType, len(body),
	)

	// HTTP headers
	connection.Write([]byte(header))
	// image body
	connection.Write(body)
}
