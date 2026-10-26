package main

import (
	"fmt"
	"httpServer/internal/http1"
	"net/http"
)

const addr = ":8080"

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	})
	if err := http1.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
