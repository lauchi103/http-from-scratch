package http1

import (
	"net/http"
	"sort"
)

type Handler interface {
	ServeHTTP(resp *Response, req *http.Request)
}

type HandlerFunc func(w ResponseWriter, req *http.Request)

type handlerFuncPattern struct {
	handlerFunc HandlerFunc
	pattern     string
}

type Mux struct {
	handlers []handlerFuncPattern //sorted list of handlers
}

var DefaultMux *Mux = NewMux()

func NewMux() *Mux {
	return &Mux{handlers: make([]handlerFuncPattern, 0)}
}

// matches the router from req to ist handlers and then executes the fitting handlerFunc
func (*Mux) ServeHTTP(resp *Response, req *http.Request) {

}

// registers the HandlerFunc to the DefaultMux and sorts them after inserting
func HandleFunc(path string, handlerFunc HandlerFunc) {
	DefaultMux.insertHandler(path, handlerFunc)
}

// sorts handlers by length
func (*Mux) insertHandler(path string, handlerFunc HandlerFunc) {
	handlerFP := handlerFuncPattern{handlerFunc: handlerFunc, pattern: path}
	DefaultMux.handlers = append(DefaultMux.handlers, handlerFP)

	sort.Slice(DefaultMux.handlers, func(i, j int) bool {
		return len(DefaultMux.handlers[i].pattern) < len(DefaultMux.handlers[j].pattern)
	})
}
