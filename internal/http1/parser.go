package http1

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var validMethods = map[string]struct{}{
	"GET":    {},
	"POST":   {},
	"PUT":    {},
	"DELETE": {},
	"HEAD":   {},
}

var validHTTPVersions = [2]string{"HTTP/1.1", "HTTP/1.0"}
var httpMinorVersionRegex = regexp.MustCompile(`^HTTP/(\d\.\d)(\.\d+)+$`)
var httpMissingMinorVersion = regexp.MustCompile(`^HTTP/(\d)+$`)
var httpCorrectVersionSyntax = regexp.MustCompile(`^HTTP/(\d+\.\d+)$`)

var headerLineRegex = regexp.MustCompile(`^([^:%\s]+):\s*([^\t\r\n]*)$`)

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	lineBytes, err := readCRLFLine(reader)
	if err != nil {
		return nil, err
	}

	lineArgs := strings.Split(string(lineBytes), " ")

	if len(lineArgs) != 3 {
		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("number of args in first line is != 3: %s", string(lineBytes)))
	}

	method := lineArgs[0]

	if _, exists := validMethods[strings.ToUpper(method)]; !exists {
		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("method is not valid: %s", string(method)))
	}

	//using regex ensure that path string is valid
	path := lineArgs[1]
	parsedURL, err := url.ParseRequestURI(path)
	if err != nil {
		return nil, fmt.Errorf("path is not valid: %v", path)
	}

	httpVersion := lineArgs[2]
	if !httpVersionIsValid(httpVersion) {
		if httpCorrectVersionSyntax.MatchString(httpVersion) {
			return nil, newHTTPError(http.StatusHTTPVersionNotSupported, fmt.Sprintf("the http version is not supported: %s", httpVersion))
		} else if httpMissingMinorVersion.MatchString(httpVersion) {
			return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("http version is missing minor: %s", httpVersion))
		} else if httpMinorVersionRegex.MatchString(httpVersion) {
			return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("http version with extra minor version: %s", httpVersion))
		}

		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("http version is not valid: %s", httpVersion))
	}

	req := &http.Request{
		Method:     method,
		URL:        parsedURL,
		RequestURI: path,
		Proto:      httpVersion,
		Header:     make(http.Header),
	}

	for {
		// too long lines will be handled later
		if len(req.Header) > 50 {
			return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("too many headers, maximum is 50 headers"))
		}

		lineBytes, err := readCRLFLine(reader)
		if err != nil {
			return nil, err
		}
		if len(lineBytes) == 0 {
			break
		}
		headerLine := string(lineBytes)
		if !headerLineRegex.MatchString(headerLine) {
			return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("bad header: %s", headerLine))
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

		if headerKey == "Content-Length" {
			if req.Header.Get(headerKey) != "" && req.Header.Get(headerKey) != headerVal {
				return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("content-length twice and mismatch"))
			}
		}

		req.Header.Set(headerKey, headerVal)

	}

	contentLengthStr := req.Header.Get("Content-Length")
	if strings.ToLower(req.Header.Get("Transfer-Encoding")) == "chunked" {
		//read every line and check if null chunk there, if not
		var bodyBuf bytes.Buffer
		for {
			// chunkheader
			lineBytes, err := readCRLFLine(reader)
			if err != nil {
				return nil, newHTTPError(http.StatusBadRequest, "unexpected EOF or bad chunk size line")
			}

			// chunklen
			chunkLenExtStr := string(lineBytes)
			chunkLenExtSl := strings.SplitN(chunkLenExtStr, ";", 2)
			chunkLen, err := strconv.ParseInt(chunkLenExtSl[0], 16, 64)
			if err != nil || chunkLen < 0 {
				return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid chunk size: %s", string(lineBytes)))
			}

			// zerochunk
			if chunkLen == 0 {
				if req.Trailer == nil {
					req.Trailer = make(http.Header)
				}
				for {
					trailerLine, err := readCRLFLine(reader)
					if err != nil {
						return nil, newHTTPError(http.StatusBadRequest, "error reading chunk trailer")
					}
					if len(trailerLine) == 0 {
						break
					}
					if headerLineRegex.MatchString(string(trailerLine)) {
						parts := strings.SplitN(string(trailerLine), ":", 2)
						key := strings.TrimSpace(parts[0])
						val := strings.TrimSpace(parts[1])

						req.Trailer.Add(key, val)
						continue
					}
					return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid trailer"))
				}
				break

			}

			// verify chunklen
			chunkData := make([]byte, chunkLen)
			if _, err := io.ReadFull(reader, chunkData); err != nil {
				return nil, newHTTPError(http.StatusBadRequest, "chunk body truncated before size reached")
			}

			// crlf after data
			crlf := make([]byte, 2)
			if _, err := io.ReadFull(reader, crlf); err != nil || crlf[0] != '\r' || crlf[1] != '\n' {
				return nil, newHTTPError(http.StatusBadRequest, "missing CRLF after chunk data")
			}

			bodyBuf.Write(chunkData)
		}

		req.Body = io.NopCloser(bytes.NewReader(bodyBuf.Bytes()))

	} else if contentLengthStr != "" {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil || contentLength < 0 {
			return nil, fmt.Errorf("ungültige Content-Length: %s", contentLengthStr)
		}

		req.ContentLength = contentLength
		bodyBytes := make([]byte, contentLength)
		_, err = io.ReadFull(reader, bodyBytes)
		if err != nil {
			return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("body shorter than Content-Length"))
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		req.ContentLength = 0
		req.Body = http.NoBody
	}

	return req, nil
}

func httpVersionIsValid(version string) bool {
	for _, v := range validHTTPVersions {
		if v == version {
			return true
		}
	}
	return false
}

func readCRLFLine(reader *bufio.Reader) ([]byte, error) {
	lineBytes, err := reader.ReadSlice('\n')
	if err != nil {
		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("bad read, CRLF is missing"))
	}
	if len(lineBytes) < 2 {
		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("line is too short: len=%d, %s", len(lineBytes), string(lineBytes)))
	} else if lineBytes[len(lineBytes)-2] != 13 || lineBytes[len(lineBytes)-1] != 10 {
		return nil, newHTTPError(http.StatusBadRequest, fmt.Sprintf("bad line ending, CRLF missing: %s", string(lineBytes)))
	}
	return lineBytes[:len(lineBytes)-2], nil
}
