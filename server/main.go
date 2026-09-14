package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	// informs the main image folder
	imagesDir = flag.String("images", "images", "Folder holding the images to serve")
	// holds the answer back on purpose, so the balancer has a slow backend to route around
	delay = flag.Duration("delay", 0, "Artificial delay before answering")
)

// makes sure the connection is up for the LD to send stuff
func main() {
	port := flag.String("port", "8081", "Port to listen on")
	flag.Parse()

	if err := os.MkdirAll(*imagesDir, 0755); err != nil {
		log.Fatalf("Failed to create images directory: %v", err)
	}

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

// handles the connection and distributes tasks based on the input
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

	switch {
	case method == "GET" && path == "/random-image":
		serveImage(connection)
	default:
		sendResponse(connection, "404 Not Found", "text/plain", []byte("Non-existing path"))
	}
}

// serves a random image from the 'images' folder
func serveImage(connection net.Conn) {
	// the connection stays open while it sleeps, which is what makes this
	// instance look busy to the balancer
	time.Sleep(*delay)

	name, err := pickRandomImage()
	if err != nil {
		sendResponse(connection, "404 Not Found", "text/plain", []byte(err.Error()))
		return
	}

	body, err := os.ReadFile(filepath.Join(*imagesDir, name))
	if err != nil {
		sendResponse(connection, "404 Not Found", "text/plain", []byte("Image not Found"))
		return
	}

	sendResponse(connection, "200 OK", contentTypeFor(name), body, "X-Image-Name: "+name)
}

func pickRandomImage() (string, error) {
	entries, err := os.ReadDir(*imagesDir)
	if err != nil {
		return "", fmt.Errorf("could not read the images folder")
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no images on the folder")
	}

	return files[rand.Intn(len(files))], nil
}

func contentTypeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func sendResponse(connection net.Conn, status, contentType string, body []byte, extraHeaders ...string) {
	header := fmt.Sprintf(
		"HTTP/1.1 %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n",
		status, contentType, len(body),
	)

	for _, extra := range extraHeaders {
		header += extra + "\r\n"
	}
	header += "Connection: close\r\n\r\n"

	// HTTP headers
	connection.Write([]byte(header))
	// body: image bytes or a plain text message
	connection.Write(body)
}
