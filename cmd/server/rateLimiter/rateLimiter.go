package rateLimiter

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// The formula for this rate limiter is:
// newToken = min (burst, lastTokens + (time * rate))
// rate: tokens/sec

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	rateLimit  float64
	burstLimit float64
	// why time function: in testing, we need to change time manually
	timeFunction func() time.Time
}

type bucket struct {
	remainingTokens    float64
	lastTimeBucketUsed time.Time
}

func NewRateLimiter(perMinute int, burst int) *RateLimiter {
	return &RateLimiter{
		buckets:      make(map[string]*bucket),
		rateLimit:    float64(perMinute) / 60.0,
		burstLimit:   float64(burst),
		timeFunction: time.Now,
	}
}

// checking each ip
func (obj *RateLimiter) Allow(ip string) (bool, time.Duration) {
	obj.mu.Lock()
	defer obj.mu.Unlock()

	now := obj.timeFunction()
	ipBucket, ok := obj.buckets[ip]
	if !ok {
		// at first, reamaining token is burst limit value
		ipBucket = &bucket{
			remainingTokens:    obj.burstLimit,
			lastTimeBucketUsed: now,
		}

		obj.buckets[ip] = ipBucket
	}

	// newToken = min (burst, lastTokens + (time * rate))
	ipBucket.remainingTokens =
		math.Min(
			obj.burstLimit,
			ipBucket.remainingTokens+now.Sub(ipBucket.lastTimeBucketUsed).Seconds()*obj.rateLimit,
		)

	// last time that ip has sent a request should be saved
	ipBucket.lastTimeBucketUsed = now

	if ipBucket.remainingTokens < 1 {
		// how much time does it need to reach exactly 1 token
		wait := time.Duration((1 - ipBucket.remainingTokens) / obj.rateLimit * float64(time.Second))
		return false, wait
	}

	ipBucket.remainingTokens--
	return true, 0
}

// remove buckets that were not used for a while (to not fill memory)
func (obj *RateLimiter) cleanup(maxLivingTime time.Duration) {
	obj.mu.Lock()
	defer obj.mu.Unlock()

	now := obj.timeFunction()
	for ip, b := range obj.buckets {
		if now.Sub(b.lastTimeBucketUsed) > maxLivingTime {
			delete(obj.buckets, ip)
		}
	}
}

func (obj *RateLimiter) CleanupLoop(ctx context.Context, maxMinute int) {

	// ticker is checking function every minutes
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	// check context and ticker channels
	for {
		select {
		case <-ctx.Done():
			return // it is for graceful shutdown
		case <-ticker.C:
			obj.cleanup(time.Duration(maxMinute) * time.Minute)
		}
	}
}

func (obj *RateLimiter) RateLimiterGateway(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// getting the ip from request
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		// checking its status and allow it to proceed
		ok, wait := obj.Allow(ip)
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			http.Error(w, "Too many requests", http.StatusTooManyRequests) // 429
			return
		}

		// proceed its path
		next.ServeHTTP(w, r)
	})
}
