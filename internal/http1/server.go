package http1

import (
	"fmt"
	"log"
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
		//panic("ListenAndServe: Not implemented")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			msg := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nHello World\n")
			c.Write(msg)
			c.Close()
		}(conn)
		/*
				Functions that might be useful:
				* listener.Accept()
			  * HandleConnection()
		*/
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	panic("HandleConnection: Not Implemented")
	/*
			Functions that might be useful:
			* ParseRequest()
		  * mux.ServeHTTP()
	*/
}
