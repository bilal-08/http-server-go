package main

import (
	"fmt"
	"strings"
)

type HttpRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type HttpResponse struct {
	Status  string
	Headers map[string]string
	Body    string
}

func NewResponse(status string, headers map[string]string, body string) HttpResponse {
	return HttpResponse{
		Status:  status,
		Headers: headers,
		Body:    body,
	}
}

func (r HttpResponse) ToString() []byte {
	headers := ""
	for key, value := range r.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %s\r\n%s\r\n%s", r.Status, headers, r.Body))
}

func ParseRequest(req []byte) HttpRequest {
	
	lines := strings.Split(string(req), "\r\n")
	firstLine := strings.Split(lines[0], " ")
	method := firstLine[0]
	path := firstLine[1]
	headers := make(map[string]string)
	body := ""
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			body = strings.Join(lines[i+1:], "\r\n")
			break
		}
		parts := strings.Split(lines[i], ": ")
		headers[parts[0]] = parts[1]
	}
	return HttpRequest{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}
}