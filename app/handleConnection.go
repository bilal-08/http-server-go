package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	
	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	req = req[:n]
	
	request := ParseRequest(req)
	fmt.Println(request)
	
	var response HttpResponse
	
	switch {
	case strings.HasPrefix(request.Path, "/echo/"):
		response = handleEcho(request)
	case request.Path == "/user-agent":
		response = handleUserAgent(request)
	case strings.HasPrefix(request.Path, "/files/"):
		if request.Method == "GET" {
			response = handleGetFile(request)
		} else if request.Method == "POST" {
			response = handlePostFile(request)
		}
	case request.Path == "/":
		response = NewResponse("200 OK", map[string]string{"Content-Length": "0"}, "")
	default:
		response = NewResponse("404 Not Found", map[string]string{"Content-Length": "0"}, "")
	}
	
	conn.Write(response.ToString())
}

func handleEcho(req HttpRequest) HttpResponse {
	echo := strings.TrimPrefix(req.Path, "/echo/")
	
	if strings.Contains(req.Headers["Accept-Encoding"], "gzip") {
		compressed, err := compressGzip([]byte(echo))
		if err != nil {
			fmt.Println("Error compressing response:", err)
			return NewResponse("500 Internal Server Error", nil, "")
		}
		
		headers := map[string]string{
			"Content-Type": "text/plain",
			"Content-Length": strconv.Itoa(len(compressed)),
			"Content-Encoding": "gzip",
		}
		return NewResponse("200 OK", headers, string(compressed))
	}
	
	headers := map[string]string{
		"Content-Type": "text/plain",
		"Content-Length": strconv.Itoa(len(echo)),
	}
	return NewResponse("200 OK", headers, echo)
}

func handleUserAgent(req HttpRequest) HttpResponse {
	userAgent := req.Headers["User-Agent"]
	if userAgent == "" {
		return NewResponse("404 Not Found", map[string]string{"Content-Length": "0"}, "")
	}
	
	headers := map[string]string{
		"Content-Type": "text/plain",
		"Content-Length": strconv.Itoa(len(userAgent)),
	}
	return NewResponse("200 OK", headers, userAgent)
}

func handleGetFile(req HttpRequest) HttpResponse {
	filename := strings.TrimPrefix(req.Path, "/files/")
	tempDir := os.Args[2]
	filepath := filepath.Join(tempDir, filename)
	
	content, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewResponse("404 Not Found", nil, "")
		}
		fmt.Println("Error reading file:", err)
		return NewResponse("500 Internal Server Error", nil, "")
	}
	
	headers := map[string]string{
		"Content-Type": "application/octet-stream",
		"Content-Length": strconv.Itoa(len(content)),
	}
	return NewResponse("200 OK", headers, string(content))
}

func handlePostFile(req HttpRequest) HttpResponse {
	filename := strings.TrimPrefix(req.Path, "/files/")
	tempDir := os.Args[2]
	filepath := filepath.Join(tempDir, filename)
	
	err := os.WriteFile(filepath, []byte(req.Body), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return NewResponse("500 Internal Server Error", nil, "")
	}
	
	return NewResponse("201 Created", nil, "")
}

func compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}