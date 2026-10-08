package storeDB

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"url-shortener/internal"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// two helper functions:

func openDB(t *testing.T, path string) (*gorm.DB, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}

	// one connection for not getting lock error in goroutines
	sqlDB.SetMaxOpenConns(1)

	closeFuntionForCleanUp := func() { sqlDB.Close() }

	t.Cleanup(closeFuntionForCleanUp)
	return db, closeFuntionForCleanUp
}

func newTestStore(t *testing.T) *SqlStore {
	t.Helper()

	db, _ := openDB(t, filepath.Join(t.TempDir(), "temp.db"))
	store, err := New(db)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	return store
}

// same tests from store_test.go

func TestCreateAndGet(t *testing.T) {
	store := newTestStore(t)
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
	if longURL.LongUrl != url {
		t.Errorf("expected %s, got %s", url, longURL)
	}
}

func TestIdempotency(t *testing.T) {
	store := newTestStore(t)
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
	store := newTestStore(t)
	_, err := store.GetLongUrl("100001")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, internal.ErrNotFound) {
		t.Errorf("expected ErrNotFound, but got: %v", err)
	}
}

func TestConcurrentRequests(t *testing.T) {
	store := newTestStore(t)
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

// new test for part 4
func TestRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restart.db")
	url := "https://salam.com/hello"

	// -------------------------------------------------
	db1, close1 := openDB(t, path)
	store1, err := New(db1)
	if err != nil {
		t.Fatalf("failed to create first store: %v", err)
	}

	code1, err := store1.GetShortCode(url)
	if err != nil {
		t.Fatalf("error on GetShortCode: %v", err)
	}
	close1()
	// -------------------------------------------------

	// -------------------------------------------------
	db2, _ := openDB(t, path)
	store2, err := New(db2)
	if err != nil {
		t.Fatalf("failed to create second store: %v", err)
	}

	longURL, err := store2.GetLongUrl(code1)
	if err != nil {
		t.Fatalf("code %s is lost after restart: %v", code1, err)
	}
	if longURL.LongUrl != url {
		t.Errorf("expected %s, got %s", url, longURL.LongUrl)
	}

	// same url, same code
	code2, err := store2.GetShortCode(url)
	if err != nil {
		t.Fatalf("error on GetShortCode after restart: %v", err)
	}
	if code1 != code2 {
		t.Errorf("idempotency after restart is not satisfied, before=%s after=%s", code1, code2)
	}
	// -------------------------------------------------

	// test to not taking only one code
	code3, err := store2.GetShortCode("https://salam.com/thirdpart")
	if err != nil {
		t.Fatalf("error on GetShortCode for new url: %v", err)
	}
	if code3 == code1 {
		t.Errorf("new url got an old code: %s", code3)
	}
}
