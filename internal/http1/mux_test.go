package http1

import (
	"net/http"
	"slices"
	"testing"
)

func TestMuxMethods(t *testing.T) {
	dummyHandler := func(w ResponseWriter, req *http.Request) {}

	tests := []struct {
		name  string
		mux   *Mux
		input []handlerFuncPattern
		want  []handlerFuncPattern
	}{
		{
			name: "unsorted patterns sorted by length descending",
			input: []handlerFuncPattern{
				{pattern: "/", handlerFunc: dummyHandler},
				{pattern: "/api/v1/users", handlerFunc: dummyHandler},
				{pattern: "/api/", handlerFunc: dummyHandler},
			},
			want: []handlerFuncPattern{
				{pattern: "/api/v1/users", handlerFunc: dummyHandler},
				{pattern: "/api/", handlerFunc: dummyHandler},
				{pattern: "/", handlerFunc: dummyHandler},
			},
			mux: NewMux(),
		},
		{
			name: "already sorted patterns",
			input: []handlerFuncPattern{
				{pattern: "/api/v1/users", handlerFunc: dummyHandler},
				{pattern: "/api/", handlerFunc: dummyHandler},
				{pattern: "/", handlerFunc: dummyHandler},
			},
			want: []handlerFuncPattern{
				{pattern: "/api/v1/users", handlerFunc: dummyHandler},
				{pattern: "/api/", handlerFunc: dummyHandler},
				{pattern: "/", handlerFunc: dummyHandler},
			},
			mux: NewMux(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, in := range tc.input {
				tc.mux.HandleFunc(in.pattern, in.handlerFunc)
			}

			// 1. Tatsächliche Patterns aus dem Mux extrahieren
			gotPatterns := make([]string, len(tc.mux.handlers))
			for i, h := range tc.mux.handlers {
				gotPatterns[i] = h.pattern
			}

			// 2. Erwartete Patterns aus tc.want extrahieren
			wantPatterns := make([]string, len(tc.want))
			for i, w := range tc.want {
				wantPatterns[i] = w.pattern
			}

			// 3. Auf Gleichheit und Reihenfolge prüfen
			if !slices.Equal(gotPatterns, wantPatterns) {
				t.Errorf("falsche Reihenfolge:\ngot:  %v\nwant: %v", gotPatterns, wantPatterns)
			}
		})
	}
}
