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

	b.ReportAllocs()
	b.ResetTimer()

	i := 0

	for b.Loop() {
		i++
		body := fmt.Sprintf(`{"url":"https://salam.com/%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("expected 201, got %d", w.Code)
		}
	}
}

func BenchmarkGetLongUrlDiff(b *testing.B) {
	mux := setupTestServer()

	b.ReportAllocs()
	b.ResetTimer()

	i := 0

	for b.Loop() {
		b.StopTimer()
		i++
		body := fmt.Sprintf(`{"url":"https://salam.com/%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("need 201, got %d", w.Code)
		}

		// .......................

		var resp CreateCodeResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			b.Fatalf("error in decodeing response: %v", err)
		}

		if resp.Code == "" || !strings.Contains(resp.ShortURL, resp.Code) {
			b.Fatalf("invalid response: %+v", resp)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
		w2 := httptest.NewRecorder()

		b.StartTimer()

		mux.ServeHTTP(w2, req2)

		if w2.Code != http.StatusFound {
			b.Fatalf("need 302, got %d", w2.Code)
		}
	}
}

func BenchmarkGetLongUrlSame(b *testing.B) {
	mux := setupTestServer()

	b.ReportAllocs()
	b.ResetTimer()

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

	for b.Loop() {
		req2 := httptest.NewRequest(http.MethodGet, "/"+resp.Code, nil)
		w2 := httptest.NewRecorder()

		mux.ServeHTTP(w2, req2)

		if w2.Code != http.StatusFound {
			b.Fatalf("need 302, got %d", w2.Code)
		}
	}
}
