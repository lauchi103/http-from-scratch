package main

import (
	"fmt"
	"httpServer/internal/http1"
	"io"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const addr = ":8080"

func main() {
	metricsHandler := promhttp.Handler()

	http1.HandleFunc("/metrics", func(w http1.ResponseWriter, r *http.Request) {
		metricsHandler.ServeHTTP(w, r)
	})

	http1.HandleFunc("/hello", func(w http1.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})
	http1.HandleFunc("/echo", func(w http1.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := io.Copy(w, r.Body)
		if err != nil {
			//http.Error(w, "Failed to copy body", http.StatusInternalServerError)
			w.WriteHeader(500)
			io.WriteString(w, fmt.Sprintf("failed to copy body: %d", http.StatusInternalServerError))
		}

	})
	//store := NewUserStorage()
	//http.HandleFunc("GET /users", handleGetUsers(store))

	//tls
	/*if err := http1.ListenAndServeTLS(":8443", nil); err != nil {
		log.Fatalf("Server-Abbruch: %v", err)
	}*/

	//tcp

	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}
