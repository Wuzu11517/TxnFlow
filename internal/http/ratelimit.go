package http

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter tracks request counts per IP address
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// Visitor tracks request count and reset time for an IP
type Visitor struct {
	count     int
	lastReset time.Time
}

// Global rate limiter instance
var limiter *RateLimiter

func init() {
	// Initialize rate limiter
	// 100 requests per hour per IP address
	limiter = &RateLimiter{
		visitors: make(map[string]*Visitor),
		limit:    100,
		window:   time.Hour,
	}

	// Start cleanup goroutine to remove old entries
	go limiter.cleanupVisitors()
}

// cleanupVisitors removes stale visitor entries every 5 minutes
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, visitor := range rl.visitors {
			// Remove visitors that haven't been seen in 2x the window
			if time.Since(visitor.lastReset) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getVisitor returns the visitor for an IP, creating one if needed
func (rl *RateLimiter) getVisitor(ip string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	visitor, exists := rl.visitors[ip]
	if !exists {
		// Create new visitor
		visitor = &Visitor{
			count:     0,
			lastReset: time.Now(),
		}
		rl.visitors[ip] = visitor
	}

	return visitor
}

// isAllowed checks if a request from this IP should be allowed
func (rl *RateLimiter) isAllowed(ip string) (bool, int, time.Time) {
	visitor := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if we need to reset the counter
	if time.Since(visitor.lastReset) > rl.window {
		visitor.count = 0
		visitor.lastReset = time.Now()
	}

	// Check if under limit
	if visitor.count >= rl.limit {
		// Over limit - return false
		resetTime := visitor.lastReset.Add(rl.window)
		return false, 0, resetTime
	}

	// Under limit - increment and allow
	visitor.count++
	remaining := rl.limit - visitor.count
	resetTime := visitor.lastReset.Add(rl.window)

	return true, remaining, resetTime
}

// RateLimitMiddleware is Chi middleware for rate limiting
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := getClientIP(r)

		// Check if request is allowed
		allowed, remaining, resetTime := limiter.isAllowed(ip)

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			// Rate limit exceeded
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
			http.Error(w, 
				fmt.Sprintf("Rate limit exceeded. Try again at %s", resetTime.Format(time.RFC1123)), 
				http.StatusTooManyRequests)
			return
		}

		// Request allowed - continue to next handler
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the real client IP from the request
// Handles proxies, load balancers, and direct connections
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	// This is important for Railway, Heroku, etc.
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2)
		// The first one is the original client
		ips := strings.Split(forwarded, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	// Check X-Real-IP header (alternative header used by some proxies)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	// This is the direct connection IP (may be a proxy if behind a load balancer)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return as-is if parsing fails
	}

	return ip
}

// GetRateLimitInfo returns current rate limit status for an IP (for debugging/monitoring)
func GetRateLimitInfo(ip string) (count int, limit int, resetTime time.Time) {
	visitor := limiter.getVisitor(ip)

	limiter.mu.RLock()
	defer limiter.mu.RUnlock()

	if time.Since(visitor.lastReset) > limiter.window {
		// Counter would be reset on next request
		return 0, limiter.limit, time.Now().Add(limiter.window)
	}

	resetTime = visitor.lastReset.Add(limiter.window)
	return visitor.count, limiter.limit, resetTime
}
