package http1

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type Response struct {
	header        http.Header
	statusCode    int
	headerWritten bool
	conn          net.Conn
	bw            *bufio.Writer
}

func NewResponse(conn net.Conn) *Response {
	return &Response{conn: conn, bw: bufio.NewWriter(conn), header: make(http.Header)}
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Write(data []byte) (n int, err error) {
	if r.headerWritten == false {
		r.WriteHeader(http.StatusOK)
	}

	n, err = r.bw.Write(data)
	if err != nil {
		return n, err
	}

	err = r.bw.Flush()
	if err != nil {
		return n, err
	}

	return n, nil

}

// error handling?
func (r *Response) WriteHeader(statusCode int) {
	if r.headerWritten {
		return
	}
	r.headerWritten = true
	r.statusCode = statusCode

	if r.header.Get("Connection") == "" {
		r.header.Set("Connection", "close")
	}

	fmt.Fprintf(r.bw, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))

	r.header.Write(r.bw)

	r.bw.WriteString("\r\n")
}
