package http1

import (
	"bufio"
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

// run different Request for testing
func TestParseRequest(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		target   string
		body     string
		headers  map[string]string
		wantErr  bool
		wantPath string
	}{
		{
			name:     "Simple GET root",
			method:   "GET",
			target:   "/",
			wantPath: "/",
		},
		{
			name:     "GET with nested path and query",
			method:   "GET",
			target:   "/api/v1/users?active=true",
			wantPath: "/api/v1/users",
		},
		{
			name:     "POST with body",
			method:   "POST",
			target:   "/submit",
			body:     `{"status":"ok"}`,
			wantPath: "/submit",
		},
		{
			name:     "Custom headers",
			method:   "GET",
			target:   "/headers",
			headers:  map[string]string{"X-Test-Trace": "12345", "Connection": "closed"},
			wantPath: "/headers",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			}

			var req = httptest.NewRequest(tc.method, tc.target, bodyReader)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			// 2. Serialize request to wire format
			var buf bytes.Buffer
			if err := req.Write(&buf); err != nil {
				t.Fatalf("Failed to write request: %v", err)
			}

			// 3. Parse via your custom parser
			reader := bufio.NewReader(&buf)
			parsedReq, err := ParseRequest(reader)

			// 4. Verify error expectations
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseRequest() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return // Don't check fields if an error was expected
			}

			// 5. Assert fields
			if parsedReq.Method != tc.method {
				t.Errorf("Method mismatch: got %q, want %q", parsedReq.Method, tc.method)
			}
			if parsedReq.URL.Path != tc.wantPath {
				t.Errorf("Path mismatch: got %q, want %q", parsedReq.URL.Path, tc.wantPath)
			}

			for k, wantVal := range tc.headers {
				if gotVal := parsedReq.Header.Get(k); gotVal != wantVal {
					t.Errorf("Header %q mismatch: got %q, want %q", k, gotVal, wantVal)
				}
			}
		})
	}
}
