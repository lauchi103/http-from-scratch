package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			err := HandleConnection(conn, handler)
			if err != nil {
				log.Printf("handleConnection with error: %v", err)
			}
		}()
		if err != nil {
			return err
		}
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		resp := NewResponse(conn)

		req, err := ParseRequest(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else if httpErr := (*HTTPError)(nil); errors.As(err, &httpErr) {
				resp.Header().Set("Connection", "close")
				resp.WriteHeader(httpErr.StatusCode)
				resp.Flush()
				return nil
			}
			return err
		}

		//check for connection alive / close
		isClose := strings.EqualFold(req.Header.Get("Connection"), "close")
		if isClose {

		} else if req.Proto == "HTTP/1.1" || (req.Proto == "HTTP/1.0" && req.Header.Get("Connection") == "keep-alive") {
			resp.Header().Set("Keep-Alive", "timeout=5, max=1000")
			resp.Header().Set("Connection", "keep-alive")
		} else {
			resp.Header().Set("Connection", "close")
		}

		mux.ServeHTTP(resp, req)
		resp.Flush()
		if isClose || strings.EqualFold(resp.Header().Get("Connection"), "close") {
			return nil
		}
	}
}
