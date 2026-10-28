package main

import (
	"fmt"
	"httpServer/internal/http1"
	"io"
	"net/http"
)

const addr = ":8080"

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := io.Copy(w, r.Body)
		if err != nil {
			http.Error(w, "Failed to copy body", http.StatusInternalServerError)
		}
	})
	store := NewUserStorage()
	http.HandleFunc("GET /users", handleGetUsers(store))

	//tls
	/*if err := http1.ListenAndServeTLS(":8443", nil); err != nil {
		log.Fatalf("Server-Abbruch: %v", err)
	}*/

	//tcp

	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}
