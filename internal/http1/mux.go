package http1

import (
	"net/http"
	"sort"
	"strings"
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
func (mux *Mux) ServeHTTP(resp *Response, req *http.Request) {
	for _, handlerFunc := range mux.handlers {
		if strings.EqualFold(handlerFunc.pattern, req.RequestURI) {
			handlerFunc.handlerFunc(resp, req)
		}
	}
}

// registers the HandlerFunc to the DefaultMux and sorts them after inserting
func HandleFunc(path string, handlerFunc HandlerFunc) {
	DefaultMux.handleFunc(path, handlerFunc)
}

func (mux *Mux) handleFunc(path string, handlerFunc HandlerFunc) {
	mux.insertHandler(path, handlerFunc)
}

// sorts handlers by length
func (m *Mux) insertHandler(path string, handlerFunc HandlerFunc) {
	handlerFP := handlerFuncPattern{handlerFunc: handlerFunc, pattern: path}
	m.handlers = append(m.handlers, handlerFP)

	sort.Slice(m.handlers, func(i, j int) bool {
		return len(m.handlers[i].pattern) > len(m.handlers[j].pattern)
	})
}
