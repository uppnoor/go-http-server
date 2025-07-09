package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	fmt.Println("Listening on http://localhost:8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	fmt.Print("Request: ", requestLine)

	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		return
	}

	method := parts[0]
	path := parts[1]

	if method != "GET" {
		writeResponse(conn, "405 Method Not Allowed", "Method Not Allowed")
		return
	}

	if path == "/" {
		serveFile(conn, "public/index.html")
	} else {
		writeResponse(conn, "404 Not Found", "Page Not Found")
	}
}

func serveFile(conn net.Conn, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		writeResponse(conn, "500 Internal Server Error", "Failed to read file")
		return
	}

	header := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(content)) +
		"\r\n"

	conn.Write([]byte(header))
	conn.Write(content)
}

func writeResponse(conn net.Conn, status string, body string) {
	header := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n", status, len(body))
	io.WriteString(conn, header+body)
}