package rateLimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	now := time.Now()

	rl := NewRateLimiter(60, 5)

	// this function allows us to change the return value any time we want
	rl.timeFunction = func() time.Time { return now }

	// burst 5 times
	for i := 0; i < 5; i++ {
		if ok, _ := rl.Allow("8.8.8.8"); !ok {
			t.Fatalf("request %d should pass", i+1)
		}
	}
	// it should block
	if ok, wait := rl.Allow("8.8.8.8"); ok || wait <= 0 {
		t.Fatal("6th request should be blocked with a positive wait")
	}

	// another IP should pass
	if ok, _ := rl.Allow("9.9.9.9"); !ok {
		t.Fatal("another ip cannot pass")
	}

	// one second later: exactly one token is back (because we set rate to 60 req/minute)
	now = now.Add(time.Second)
	if ok, _ := rl.Allow("8.8.8.8"); !ok {
		t.Fatal("should pass after a second")
	}
	if ok, _ := rl.Allow("8.8.8.8"); ok {
		t.Fatal("unexpected passing!")
	}

	// a long pause: check it does not reached above burst limit
	now = now.Add(time.Hour * 24)
	passed := 0
	for i := 0; i < 10; i++ {
		if ok, _ := rl.Allow("8.8.8.8"); ok {
			passed++
		}
	}
	if passed != 5 {
		t.Fatalf("burst should cap at 5, got %d", passed)
	}
}

func TestCleanup(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(60, 3)
	rl.timeFunction = func() time.Time { return now }
	rl.Allow("8.8.8.4")

	now = now.Add(time.Hour)
	rl.cleanup(10 * time.Minute)
	if len(rl.buckets) != 0 {
		t.Fatal("idle bucket should be removed")
	}
}

func TestRateLimiterGateway(t *testing.T) {
	rl := NewRateLimiter(60, 3)

	// assign handler to ratelimiter gateway
	gatewayHandler := rl.RateLimiterGateway(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// send to assigned handler (getwayHandler)
	send := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)
		req.RemoteAddr = "https://salam.com:1234"
		rec := httptest.NewRecorder()
		gatewayHandler.ServeHTTP(rec, req)
		return rec
	}

	send()
	send()
	send()

	rec := send() // after 3 times , 4th req should be blocked
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("need code 429, but got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("empty Retry-After.")
	}
}
