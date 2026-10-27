package http1

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
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

func ListenAndServeTLS(addr string, handler http.Handler) error {
	cert, err := generateDummyCert()
	if err != nil {
		return fmt.Errorf("create cert: %w", err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer tcpListener.Close()

	tlsListener := tls.NewListener(tcpListener, config)
	defer tlsListener.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	for {
		conn, err := tlsListener.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			err := HandleConnection(c, handler)
			if err != nil {
				log.Printf("handleConnection with error: %v", err)
			}
		}(conn)
	}
}

func generateDummyCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour), // 1 Tag gültig
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}
