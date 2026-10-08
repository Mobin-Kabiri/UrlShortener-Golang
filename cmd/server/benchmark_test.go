package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkShortenUrl(b *testing.B) {
	mux := setupTestServer()

	// build everything before the timed section
	reqs := make([]*http.Request, b.N)
	recs := make([]*httptest.ResponseRecorder, b.N)
	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf(`{"url":"https://salam.com/%d"}`, i)
		reqs[i] = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		recs[i] = httptest.NewRecorder()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(recs[i], reqs[i])

		if recs[i].Code != http.StatusCreated {
			b.Fatalf("expected 201, got %d", recs[i].Code)
		}
	}
}
func BenchmarkGetLongUrlDiff(b *testing.B) {
	mux := setupTestServer()

	// create 10000 short codes and pre-build a GET request for each
	const numCodes = 10000
	reqs := make([]*http.Request, numCodes)
	for i := range reqs {
		body := fmt.Sprintf(`{"url":"https://salam.com/%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("need 201, got %d", w.Code)
		}

		var resp CreateCodeResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			b.Fatalf("error decoding response: %v", err)
		}
		reqs[i] = httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
	}

	b.ReportAllocs()

	i := 0
	for b.Loop() {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, reqs[i%numCodes])

		if w.Code != http.StatusFound {
			b.Fatalf("need 302, got %d", w.Code)
		}
		i++
	}
}

func BenchmarkGetLongUrlSame(b *testing.B) {
	mux := setupTestServer()

	

	body := `{"url":"https://salam.com/"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		b.Fatalf("need 201, got %d", w.Code)
	}

	var resp CreateCodeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		b.Fatalf("error in decodeing response: %v", err)
	}

	if resp.Code == "" || !strings.Contains(resp.ShortURL, resp.Code) {
		b.Fatalf("invalid response: %+v", resp)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
	
	b.ReportAllocs()
	for b.Loop() {
		
		w2 := httptest.NewRecorder()

		mux.ServeHTTP(w2, req2)

		if w2.Code != http.StatusFound {
			b.Fatalf("need 302, got %d", w2.Code)
		}
	}
}
