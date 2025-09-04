package main

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net"
	"os"
	"path/filepath"
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

	// Read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	fmt.Print("Request: ", requestLine)

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) < 3 {
		return
	}
	method := parts[0]
	path := parts[1]

	// Only support GET for now.
	if method != "GET" {
		writeResponse(conn, "405 Method Not Allowed", "text/plain", []byte("Method Not Allowed"))
		return
	}

	// routing
	switch path {
	case "/":
		serveFile(conn, filepath.Join("public", "index.html"))
		return
	case "/about":
		serveFile(conn, filepath.Join("public", "about.html"))
		return
	default:
		// Try to serve any file under ./public for other paths
		if serveStaticUnderPublic(conn, path) {
			return
		}
		writeResponse(conn, "404 Not Found", "text/plain", []byte("Page Not Found"))
		return
	}
}

// Attempts to serve a file under ./public that matches the URL path.
func serveStaticUnderPublic(conn net.Conn, urlPath string) bool {
	// Prevent directory traversal and map URL to filesystem
	clean := filepath.Clean(urlPath)
	// remove leading "/" so Join doesn't treat it as absolute
	if strings.HasPrefix(clean, "/") {
		clean = clean[1:]
	}
	// final path = ./public/<clean>
	full := filepath.Join("public", clean)

	// Ensure the resolved path is still inside ./public
	publicAbs, _ := filepath.Abs("public")
	fullAbs, _ := filepath.Abs(full)
	if !strings.HasPrefix(fullAbs, publicAbs) {
		return false
	}

	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		return false
	}
	serveFile(conn, full)
	return true
}

func serveFile(conn net.Conn, filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		writeResponse(conn, "500 Internal Server Error", "text/plain", []byte("Failed to read file"))
		return
	}

	ct := contentTypeFromExt(filename)
	writeResponse(conn, "200 OK", ct, content)
}

func writeResponse(conn net.Conn, status, contentType string, body []byte) {
	header := fmt.Sprintf(
		"HTTP/1.1 %s\r\nContent-Length: %d\r\nContent-Type: %s\r\nConnection: close\r\n\r\n",
		status, len(body), contentType,
	)
	io.WriteString(conn, header)
	conn.Write(body)
}

func contentTypeFromExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "application/octet-stream"
	}
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		if ext == ".txt" || ext == ".log" || ext == ".md" {
			return "text/plain; charset=utf-8"
		}
		return "application/octet-stream"
	}
	// ensure text types declare utf-8
	if strings.HasPrefix(ct, "text/") && !strings.Contains(ct, "charset=") {
		return ct + "; charset=utf-8"
	}
	return ct
}
