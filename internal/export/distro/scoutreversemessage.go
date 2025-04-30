// Copyright (c) 2025 Circutor S.A. All rights reserved.

package distro

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPMessage struct {
	RequestID string
	Method    string
	Path      string
	Headers   http.Header
	Body      []byte
}

type HTTPResponse struct {
	RequestID     string
	StatusCode    int
	StatusMessage string
	Headers       http.Header
	Body          []byte
}

func (m *HTTPMessage) ToHTTPRequest(baseURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), m.Method, baseURL+m.Path, bytes.NewReader(m.Body))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header = m.Headers

	return req, nil
}

func ParseHTTPRequest(rawRequest string) (*HTTPMessage, error) {
	reader := bufio.NewReader(strings.NewReader(rawRequest))

	// Read request line
	// The request line is the first line of the request and contains the method and path
	method, path, err := getMethodAndPath(reader)
	if err != nil {
		return nil, err
	}

	// Read headers
	// Headers are separated by new lines and end with an empty line
	headers, err := getHeaders(reader)
	if err != nil {
		return nil, err
	}

	// Read body
	// The body is the remaining part of the request
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	return &HTTPMessage{
		Method:  method,
		Path:    path,
		Headers: headers,
		Body:    body,
	}, nil
}

func getMethodAndPath(reader *bufio.Reader) (string, string, error) {
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("error reading request line: %w", err)
	}

	requestLine = strings.TrimSpace(requestLine)

	parts := strings.Split(requestLine, " ")

	switch len(parts) {
	case 1:
		// Only the method is present, assume path is "/"
		return parts[0], "/", nil
	case 2:
		// Method and path are present
		return parts[0], parts[1], nil
	default:
		// Invalid request line
		return "", "", fmt.Errorf("invalid request line: %s", requestLine)
	}
}

func getHeaders(reader *bufio.Reader) (http.Header, error) {
	headers := make(http.Header)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		headers.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	return headers, nil
}

func (r *HTTPResponse) FormatHTTPResponse() string {
	var b strings.Builder

	// Status line
	fmt.Fprintf(&b, "HTTP/1.1 %d %s\n", r.StatusCode, r.StatusMessage)

	// Headers
	for key, values := range r.Headers {
		for _, value := range values {
			fmt.Fprintf(&b, "%s: %s\n", key, value)
		}
	}

	// Empty line separating headers and body
	b.WriteString("\n")

	// Body
	b.Write(r.Body)

	return b.String()
}
