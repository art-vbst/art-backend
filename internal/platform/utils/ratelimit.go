package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type IPRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limit:  limit,
		window: window,
		hits:   make(map[string][]time.Time),
	}
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		now := time.Now()
		cutoff := now.Add(-l.window)

		l.mu.Lock()
		timestamps := l.hits[ip]
		// prune old
		i := 0
		for _, t := range timestamps {
			if t.After(cutoff) {
				timestamps[i] = t
				i++
			}
		}
		timestamps = timestamps[:i]
		if len(timestamps) >= l.limit {
			l.mu.Unlock()
			w.Header().Set("Retry-After", l.window.String())
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		timestamps = append(timestamps, now)
		l.hits[ip] = timestamps
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
