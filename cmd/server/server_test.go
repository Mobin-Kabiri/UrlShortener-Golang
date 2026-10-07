package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"url-shortener/fakeStore"
	"url-shortener/internal"
	"url-shortener/store"
)

// fixed: calling real function instead of recreating it
func setupTestServer() *http.ServeMux {
	srv := server{
		store:   store.New(),
		baseUrl: "http://localhost:8080",
	}
	return setupServer(&srv)
}

func setupTestServerForMaxUrl() *http.ServeMux {
	srv := server{
		store:   store.NewForTesting(),
		baseUrl: "http://localhost:8080",
	}
	return setupServer(&srv)
}

// PART 1 handling:
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

func TestErrorTablePart1(t *testing.T) {
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

// PART 2 handling:

func TestNotFoundAndInvalidURL(t *testing.T) {
	mux := setupTestServer()

	testsTable := []struct {
		name               string
		method             string
		path               string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "Not Found Error ",
			method:             http.MethodGet,
			path:               "/api/v1/links/10",
			body:               `{}`,
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:               "Invalid URL Error",
			method:             http.MethodPost,
			path:               "/api/shorten",
			body:               `{"url":"http://sal^^am.com/"}`,
			expectedStatusCode: http.StatusBadRequest,
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

func TestMaxUrlError(t *testing.T) {
	mux := setupTestServerForMaxUrl()

	testsTable := []struct {
		name               string
		method             string
		path               string
		body               string
		expectedStatusCode int
	}{
		{
			name:               "Max Url Reached",
			method:             http.MethodPost,
			path:               "/api/shorten",
			body:               `{"url":"http://salam.com/"}`,
			expectedStatusCode: http.StatusServiceUnavailable,
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

func TestLongUrlResponseFormat(t *testing.T) {
	mux := setupTestServer()

	url := "https://salam.com/#sssss"
	// endpoint: POST /api/shorten
	body := `{"url":"` + url + `"}`
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

	// endpoint: GET /api/v1/links/{code}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+resp.Code, nil)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("need 200, but got %d", getW.Code)
	}

	var resGet GetUrlObjResponse
	if err := json.NewDecoder(getW.Body).Decode(&resGet); err != nil {
		t.Fatalf("error in decodeing response: %v", err)
	}

	// time format
	_, err := time.Parse(time.RFC3339, resGet.CreatedAt)
	if err != nil {
		t.Fatalf("created_at is not RFC3339: %q (%v)", resGet.CreatedAt, err)
	}
	normalizedUrl := "https://salam.com/"
	if normalizedUrl != resGet.LongUrl {
		t.Fatalf("normalized url is not same as response url")
	}
}

func TestIdempotency(t *testing.T) {
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

	// endpoint: GET /api/v1/links/{code}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+resp.Code, nil)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("need 200, but got %d", getW.Code)
	}

	var resGet GetUrlObjResponse
	if err := json.NewDecoder(getW.Body).Decode(&resGet); err != nil {
		t.Fatalf("error in decodeing response: %v", err)
	}

	// waiting for a second
	time.Sleep(time.Second)
	// ----------------------------------------------

	// endpoint: POST /api/shorten -- adding /#salam to check normalization too
	body2 := `{"url":"https://go.dev/doc/#salam"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body2))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("need 201 Created, but got %d", w2.Code)
	}

	var resp2 CreateCodeResponse
	if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
		t.Fatalf("error in decodeing response: %v", err)
	}

	if resp2.Code == "" || !strings.Contains(resp2.ShortURL, resp2.Code) {
		t.Fatalf("invalid response: %+v", resp2)
	}
	// ----------------------------------------------

	// endpoint: GET /api/v1/links/{code}
	getReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/links/"+resp2.Code, nil)
	getW2 := httptest.NewRecorder()
	mux.ServeHTTP(getW2, getReq2)

	if getW2.Code != http.StatusOK {
		t.Fatalf("need 200, but got %d", getW2.Code)
	}

	var resGet2 GetUrlObjResponse
	if err := json.NewDecoder(getW2.Body).Decode(&resGet2); err != nil {
		t.Fatalf("error in decodeing response: %v", err)
	}

	// --------------------------------------------

	if resGet != resGet2 || resp2 != resp {
		t.Fatalf("Idempotency is not satisfied")
	}

}

func TestFakeStoreErrors(t *testing.T) {
	validBody := `{"url":"https://salam.com"}`

	tests := []struct {
		name   string
		fake   *fakestore.FakeStore
		method string
		path   string
		body   string
		want   int
	}{

		// POST api/shorten
		{
			name:   "POST /api/shorten - ErrMaxUrlReached",
			fake:   &fakestore.FakeStore{Err: internal.ErrMaxUrlReached},
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   validBody,
			want:   http.StatusServiceUnavailable, // 503
		},
		{
			name:   "POST /api/shorten - ErrInvalidURL",
			fake:   &fakestore.FakeStore{Err: fmt.Errorf("%w: test invalid url", internal.ErrInvalidURL)},
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   validBody,
			want:   http.StatusBadRequest, // 400
		},
		{
			name:   "POST /api/shorten - Internal Error",
			fake:   &fakestore.FakeStore{Err: errors.New("test error")},
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   validBody,
			want:   http.StatusInternalServerError, // 500
		},
		{
			name:   "POST /api/shorten - Success",
			fake:   &fakestore.FakeStore{Err: nil},
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   validBody,
			want:   http.StatusCreated, // 201
		},

		// GET /{code}
		{
			name:   "GET /salam - ErrUrlNotFound",
			fake:   &fakestore.FakeStore{Err: internal.ErrNotFound},
			method: http.MethodGet,
			path:   "/salam",
			want:   http.StatusNotFound, // 404
		},
		{
			name:   "GET /salam - Internal Error",
			fake:   &fakestore.FakeStore{Err: errors.New("test error")},
			method: http.MethodGet,
			path:   "/salam",
			want:   http.StatusInternalServerError, // 500
		},
		{
			name:   "GET /salam - Success",
			fake:   &fakestore.FakeStore{Err: nil},
			method: http.MethodGet,
			path:   "/salam",
			want:   http.StatusFound, // 302
		},

		// GET /api/v1/links/{code}
		{
			name:   "GET /api/v1/links/codemode - ErrUrlNotFound",
			fake:   &fakestore.FakeStore{Err: internal.ErrNotFound},
			method: http.MethodGet,
			path:   "/api/v1/links/codemode",
			want:   http.StatusNotFound, // 404
		},
		{
			name:   "GET /api/v1/links/codemode - Internal Error",
			fake:   &fakestore.FakeStore{Err: errors.New("test error")},
			method: http.MethodGet,
			path:   "/api/v1/links/codemode",
			want:   http.StatusInternalServerError, // 500
		},
		{
			name:   "GET /api/v1/links/codemode - Success",
			fake:   &fakestore.FakeStore{Err: nil},
			method: http.MethodGet,
			path:   "/api/v1/links/codemode",
			want:   http.StatusOK, // 200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// creating fake store
			srv := server{
				store:   tt.fake,
				baseUrl: "http://localhost:8080",
			}
			mux := setupServer(&srv)

			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("need status %d, but got %d", tt.want, w.Code)
			}
		})
	}
}
