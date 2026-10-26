package http1

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var validMethods = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"PUT":    {},
	"DELETE": {},
}

const allowedHTTPVersion = "HTTP/1.1"

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	lineBytes, _, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}

	lineArgs := strings.Split(string(lineBytes), " ")

	if len(lineArgs) != 3 {
		return nil, fmt.Errorf("firstline of request should have 3 args, has: %d", len(lineArgs))
	}

	method := lineArgs[0]
	if _, exists := validMethods[method]; !exists {
		return nil, fmt.Errorf("http-method not supported: %v", method)
	}

	//using regex ensure that path string is valid
	path := lineArgs[1]
	parsedURL, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, fmt.Errorf("path is not valid: %v", path)
	}

	httpVersion := lineArgs[2]
	if httpVersion != allowedHTTPVersion {
		return nil, fmt.Errorf("http version is not allowed: %v", httpVersion)
	}

	req := &http.Request{
		Method:     method,
		URL:        parsedURL,
		RequestURI: path,
		Proto:      allowedHTTPVersion,
		Header:     make(http.Header),
	}

	for {
		// too long lines will be handled later
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if len(lineBytes) == 0 {
			break
		}
		keyValue := strings.SplitN(string(lineBytes), ":", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("ungültige Header-Zeile (kein Doppelpunkt): %s", string(lineBytes))
		}

		headerKey := keyValue[0]
		if strings.TrimRight(headerKey, " \t") != headerKey {
			return nil, fmt.Errorf("ungültige Leerzeichen vor Doppelpunkt im Header: %s", string(lineBytes))
		}
		headerVal := strings.TrimSpace(keyValue[1])

		req.Header.Set(headerKey, headerVal)

	}
	// noch body parsen
	return req, nil
}
