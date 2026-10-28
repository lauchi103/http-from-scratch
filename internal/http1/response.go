package http1

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Response struct {
	header        http.Header
	statusCode    int
	headerWritten bool
	conn          net.Conn
	bw            ResponseWriter
	body          bytes.Buffer
}

type ResponseWriter = *bufio.Writer

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

	n, err = r.body.Write(data)
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
}

func (r *Response) Flush() {
	r.header.Set("Content-Length", fmt.Sprint(r.body.Len()))
	if r.headerWritten == false {
		r.WriteHeader(http.StatusOK)
	}

	fmt.Fprintf(r.bw, "HTTP/1.1 %d %s\r\n", r.statusCode, http.StatusText(r.statusCode))

	r.header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	r.header.Write(r.bw)

	r.bw.WriteString("\r\n")

	r.bw.Write(r.body.Bytes())
	r.bw.Flush()
}
