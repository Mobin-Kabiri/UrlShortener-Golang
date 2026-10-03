package store

import (
	"errors"
	"sync"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	store := New()
	url := "https://salam.com/hello"

	code, err := store.GetShortCode(url)
	if err != nil {
		t.Fatalf("error on GetShortCode: %v", err)
	}

	if len(code) < 6 || len(code) > 8 {
		t.Errorf("code length shoul be 6-8, but got %d (%s)", len(code), code)
	}

	longURL, err := store.GetLongUrl(code)
	if err != nil {
		t.Fatalf("error on GetLongUrl: %v", err)
	}

	// check input and url from store
	if longURL != url {
		t.Errorf("expected %s, got %s", url, longURL)
	}
}

func TestIdempotency(t *testing.T) {
	store := New()
	url := "https://salam.com/hello"

	code1, err := store.GetShortCode(url)
	if err != nil {
		t.Fatalf("error on first GetShortCode: %v", err)
	}

	code2, err := store.GetShortCode(url)
	if err != nil {
		t.Fatalf("error on second GetShortCode: %v", err)
	}

	if code1 != code2 {
		t.Errorf("idempotency is not satisfied, first=%s second=%s", code1, code2)
	}
}

func TestNotFoundUrl(t *testing.T) {
	store := New()
	_, err := store.GetLongUrl("somethingDoesNotExist")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUrlNotFound) {
		t.Errorf("expected ErrUrlNotFound, but got: %v", err)
	}
}

func TestConcurrentRequests(t *testing.T) {
	store := New()
	url := "https://salam.com/hello"
	numOfRoutines := 300

	preCounter := store.counter
	var wg sync.WaitGroup
	codes := make([]string, numOfRoutines)

	wg.Add(numOfRoutines)
	for i := 0; i < numOfRoutines; i++ {
		go func(index int) {
			defer wg.Done()
			code, err := store.GetShortCode(url)
			if err != nil {
				t.Errorf("routine number %d crashed, error: %v", index, err)
				return
			}
			codes[index] = code
		}(i)
	}

	// wait for all
	wg.Wait()

	for index, c := range codes {
		if c != codes[0] {
			t.Fatalf("problem in concurrency: routine %d got %s instead of %s", index, c, codes[0])
		}
	}

	currCounter := store.counter
	if currCounter-1 != preCounter {
		t.Fatal("problem in concurrency: counter has increased more than once")
	}
}
