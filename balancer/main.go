package main

import (
	"bufio"       // wraps stream readers and writers into a buffer
	"fmt"         // standard library for cli output
	"io"          // copies bytes straight from one connection into another
	"log"         // standard library for console output regarding concurrency and errors
	"math/rand"   // breaks the tie between backends that are equally idle
	"net"         // standard library for TCP/IP connections and domain sockets
	"os"          // read files from disk
	"sort"        // orders the backends by how busy they are
	"strings"     // manipulates strings in Golang
	"sync/atomic" // allows Atomic types that make it better for goroutines to read
	"time"        // health check interval and dial timeouts
)

type backend struct {
	address  string
	alive    atomic.Bool
	inflight atomic.Int64
}

var pool = []*backend{
	{address: "localhost:8081"},
	{address: "localhost:8082"},
	{address: "localhost:8083"},
	{address: "localhost:8084"},
	{address: "localhost:8085"},
}

// start the TCP listener and 'connect' to the user
func main() {
	startHealthChecks(3 * time.Second)

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

// maintain the individual connection every time it's requested
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
	case path == "/random-image":
		proxyRequest(connection, method, path) // handed over to the backend server
	default:
		sendResponse(connection, "404 Not Found", "text/plain", "404 Not Found")
	}
}

// options available of specific services
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

// probes the whole pool once, then keeps probing it on every tick
func startHealthChecks(interval time.Duration) {
	for _, b := range pool {
		probe(b) // one pass up front, so the first request isn't a coin toss
	}

	go func() {
		for range time.Tick(interval) {
			for _, b := range pool {
				go probe(b)
			}
		}
	}()
}

// a backend that accepts a TCP connection is considered alive
func probe(b *backend) {
	connection, err := net.DialTimeout("tcp", b.address, time.Second)
	if err != nil {
		// Swap reports the old value, so this only logs on the way down
		if b.alive.Swap(false) {
			log.Printf("Backend %s went down: %v", b.address, err)
		}
		return
	}
	connection.Close()

	if !b.alive.Swap(true) {
		log.Printf("Backend %s is back up", b.address)
	}
}

// hands the request to the least busy healthy backend, falling through to the
// next one whenever a backend can't be reached at all
func proxyRequest(client net.Conn, method, path string) {
	var alive []*backend
	for _, b := range pool {
		if b.alive.Load() {
			alive = append(alive, b)
		}
	}

	// a stale flag shouldn't turn a working backend into a 502, so when the
	// health checks think everything is down the whole pool is tried anyway
	if len(alive) == 0 {
		alive = append(alive, pool...)
	}

	// requests here are short, so the counters are usually all zero; shuffling
	// before a stable sort spreads those ties instead of every request picking
	// the same backend
	rand.Shuffle(len(alive), func(i, j int) { alive[i], alive[j] = alive[j], alive[i] })
	sort.SliceStable(alive, func(i, j int) bool {
		return alive[i].inflight.Load() < alive[j].inflight.Load()
	})

	for _, b := range alive {
		if relayTo(b, client, method, path) {
			return
		}
	}

	sendResponse(client, "502 Bad Gateway", "text/plain", "No backend available")
}

// forwards the request to one backend and copies its answer back, untouched.
// reports false when nothing was written to the client, so the caller can retry
func relayTo(b *backend, client net.Conn, method, path string) bool {
	backendConn, err := net.Dial("tcp", b.address)
	if err != nil {
		log.Printf("Backend %s is down: %v", b.address, err)
		b.alive.Store(false) // the next probe puts it back once it answers again
		return false
	}
	defer backendConn.Close()

	b.inflight.Add(1)
	defer b.inflight.Add(-1)

	// rebuilds the request line by hand, the same way the backend parses it
	_, err = fmt.Fprintf(backendConn,
		"%s %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Connection: close\r\n\r\n",
		method, path, b.address,
	)
	if err != nil {
		log.Printf("Failed to send the request to %s: %v", b.address, err)
		return false
	}

	log.Printf("%s -> %s", path, b.address)

	// the backend answers with a full HTTP response, so it goes back verbatim.
	// past this point bytes are already on the client, so there is no retrying
	if _, err := io.Copy(client, backendConn); err != nil {
		log.Printf("Failed to relay the backend answer: %v", err)
	}

	return true
}

// formats the response to follow the HTTP design
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
