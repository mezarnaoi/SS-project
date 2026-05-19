package routes

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// ipLimiter tracks request timestamps per IP using a sliding window.
type ipLimiter struct {
	mu       sync.Mutex
	windows  map[string][]time.Time
	limit    int
	window   time.Duration
	lastClean time.Time
}

func newIPLimiter(limit int, window time.Duration) *ipLimiter {
	return &ipLimiter{
		windows:   make(map[string][]time.Time),
		limit:     limit,
		window:    window,
		lastClean: time.Now(),
	}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Keep only timestamps inside the current window
	timestamps := l.windows[ip]
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= l.limit {
		l.windows[ip] = valid
		return false
	}

	l.windows[ip] = append(valid, now)

	// Periodic cleanup to prevent memory growth from stale IPs
	if now.Sub(l.lastClean) > 5*time.Minute {
		for k, ts := range l.windows {
			if len(ts) == 0 || ts[len(ts)-1].Before(cutoff) {
				delete(l.windows, k)
			}
		}
		l.lastClean = now
	}

	return true
}

// Exceeding the limit returns HTTP 429 Too Many Requests.
func withRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := newIPLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			if !limiter.allow(ip) {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
