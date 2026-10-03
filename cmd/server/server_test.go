package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener/internal/store"
)

func setupTestServer() *http.ServeMux {
	srv := server{
		store:   store.New(),
		baseUrl: "http://localhost:8080",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/shorten", srv.getCode)
	mux.HandleFunc("GET /{code}", srv.getLongUrl)
	return mux
}

func TestCreateAndRedirect(t *testing.T) {
	mux := setupTestServer()

	// endpoint: POST /api/shorten
	body := `{"url":"https://go.dev/doc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("need 201 Created, but got %d", w.Code)
	}

	var resp CreateCodeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error in decodeing response: %v", err)
	}

	if resp.Code == "" || !strings.Contains(resp.ShortURL, resp.Code) {
		t.Fatalf("invalid response: %+v", resp)
	}
	// ----------------------------------------------

	// endpoint: GET /{code}
	getReq := httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusFound {
		t.Fatalf("need 302, but got %d", getW.Code)
	}

	location := getW.Header().Get("Location")
	if location != "https://go.dev/doc" {
		t.Errorf("location must be https://go.dev/doc, got %s", location)
	}
	// ----------------------------------------------
}

func TestErrorTable(t *testing.T) {
	mux := setupTestServer()

	testsTable := []struct {
		name               string
		method             string
		path               string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "Empty JSON Body",
			method:             http.MethodPost,
			path:               "/api/shorten",
			body:               `{}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "Invalid JSON Format",
			method: http.MethodPost,
			path:   "/api/shorten",
			body: `{
    			"sssurl":"https://go.dev/doc/"
			}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "Empty URL",
			method: http.MethodPost,
			path:   "/api/shorten",
			body: `{
    			"url":""
			}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "bad URL",
			method: http.MethodPost,
			path:   "/api/shorten",
			body: `{
    			"url":"sala.com"
			}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "Bad Protocol",
			method: http.MethodPost,
			path:   "/api/shorten",
			body: `{
    			"url":"httsps://go.dev/doc/"
			}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:   "Normal Scenario",
			method: http.MethodPost,
			path:   "/api/shorten",
			body: `{
    			"url":"https://go.dev/doc/"
			}`,
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "Invalid Method",
			method:             http.MethodGet,
			path:               "/api/shorten",
			body:               `{}`,
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Not Found Code",
			method:             http.MethodGet,
			path:               "/sallllaaam",
			body:               "",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range testsTable {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("need status %d, but got %d", tt.expectedStatusCode, w.Code)
			}
		})
	}
}
