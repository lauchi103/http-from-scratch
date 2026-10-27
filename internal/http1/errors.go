package http1

import "fmt"

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error %d: %s", e.StatusCode, e.Message)
}

func newHTTPError(statusCode int, msg string) error {
	return &HTTPError{StatusCode: statusCode, Message: msg}
}
