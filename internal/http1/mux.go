package http1

import (
	"net/http"
	"sort"
	"strings"
)

type Handler interface {
	ServeHTTP(ResponseWriter, *http.Request)
}

type HandlerFuncType func(w ResponseWriter, req *http.Request)

func (h HandlerFuncType) ServeHTTP(w ResponseWriter, req *http.Request) {
	h(w, req)
}

type handlerFuncPattern struct {
	handlerFunc HandlerFuncType
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
func (mux *Mux) ServeHTTP(r ResponseWriter, req *http.Request) {
	for _, handlerFunc := range mux.handlers {
		if strings.EqualFold(handlerFunc.pattern, req.RequestURI) {
			handlerFunc.handlerFunc(r, req)
		}
	}
}

// registers the HandlerFuncType to the DefaultMux and sorts them after inserting
func HandleFunc(path string, handlerFunc HandlerFuncType) {
	DefaultMux.HandleFunc(path, handlerFunc)
}

func (mux *Mux) HandleFunc(path string, handlerFunc HandlerFuncType) {
	mux.insertHandler(path, handlerFunc)
}

// sorts handlers by length
func (m *Mux) insertHandler(path string, handlerFunc HandlerFuncType) {
	handlerFP := handlerFuncPattern{handlerFunc: handlerFunc, pattern: path}
	m.handlers = append(m.handlers, handlerFP)

	sort.Slice(m.handlers, func(i, j int) bool {
		return len(m.handlers[i].pattern) > len(m.handlers[j].pattern)
	})
}
