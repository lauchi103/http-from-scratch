package http1

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
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

		go HandleConnection(conn, handler)
		if err != nil {
			return err
		}
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	/*
			Functions that might be useful:
			* ParseRequest()
		  * mux.ServeHTTP()
	*/
	defer conn.Close()
	req, err := ParseRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}

	mux.ServeHTTP(NewResponse(conn), req)
	return nil
}
