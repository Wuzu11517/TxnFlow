# Rate Limiting Documentation

## Overview

TxnFlow implements IP-based rate limiting to prevent API abuse and ensure fair usage across all clients.

---

## Configuration

**Default Limits**:
- **100 requests per hour** per IP address
- Applies to **all endpoints** (GET, POST, etc.)
- Automatic cleanup of old visitor records

**Implementation**: In-memory rate limiting (suitable for single-server demos)

---

## How It Works

### 1. Request Tracking

Each request is tracked by the client's IP address:

```
Client IP: 1.2.3.4
Request 1:  ✅ Allowed (1/100)
Request 2:  ✅ Allowed (2/100)
...
Request 100: ✅ Allowed (100/100)
Request 101: 🚫 Rate Limited (429)
```

### 2. Time Window

- **Window**: 1 hour (rolling)
- **Reset**: Automatic after the hour expires
- **Tracking**: Per IP address

Example:
```
10:00 AM: Client makes request #1
10:30 AM: Client makes request #50
11:00 AM: Counter resets to 0
11:01 AM: Client can make 100 more requests
```

### 3. IP Detection

The system intelligently detects client IPs even behind proxies:

**Priority order**:
1. `X-Forwarded-For` header (first IP in the list)
2. `X-Real-IP` header
3. Direct connection IP (`RemoteAddr`)

This ensures rate limiting works correctly on platforms like Railway, Heroku, Vercel, etc.

---

## HTTP Response Headers

Every response includes rate limit information:

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1708212000
```

**Headers explained**:
- `X-RateLimit-Limit`: Maximum requests allowed per window (100)
- `X-RateLimit-Remaining`: Requests remaining in current window (87)
- `X-RateLimit-Reset`: Unix timestamp when limit resets (1708212000)

### When Rate Limited

```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1708212000
Retry-After: 3600

Rate limit exceeded. Try again at Sat, 17 Feb 2026 12:00:00 GMT
```

**Additional headers**:
- `Retry-After`: Seconds until you can retry (3600 = 1 hour)

---

## Testing Rate Limits

### Manual Test (curl)

```bash
# Make a request and check headers
curl -i http://localhost:8080/stats

# Look for:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 99
```

### Automated Test Script

```bash
# Run the test script
./scripts/test-ratelimit.sh

# This will:
# 1. Send 110 requests
# 2. First 100 should succeed (201)
# 3. Last 10 should fail (429)
# 4. Display rate limit headers
```

**Expected output**:
```
Request 1: ✅ Success (201)
Request 2: ✅ Success (201)
...
Request 100: ✅ Success (201)
Request 101: 🚫 Rate limited (429) - LIMIT REACHED!
...

Test Results
============
Total Requests:     110
Successful:         100
Rate Limited:       10

✅ Rate limiting is working correctly!
```

---

## How Frontend Should Handle Rate Limits

### JavaScript Example

```javascript
async function submitTransaction(hash) {
    try {
        const response = await fetch('https://api.txnflow.com/transactions', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                transaction_hash: hash,
                chain_id: 1
            })
        });

        // Check for rate limit
        if (response.status === 429) {
            // Get reset time from headers
            const resetTime = response.headers.get('X-RateLimit-Reset');
            const resetDate = new Date(resetTime * 1000);
            
            showError(
                `Rate limit exceeded. You can try again at ${resetDate.toLocaleTimeString()}`
            );
            return;
        }

        // Check remaining requests
        const remaining = response.headers.get('X-RateLimit-Remaining');
        if (remaining < 10) {
            showWarning(`Only ${remaining} requests remaining this hour`);
        }

        if (!response.ok) {
            throw new Error('Request failed');
        }

        const data = await response.json();
        showSuccess('Transaction submitted!');
        return data;

    } catch (error) {
        showError(error.message);
    }
}
```

### User Experience Best Practices

**DO**:
- ✅ Show clear error messages when rate limited
- ✅ Display when the user can retry
- ✅ Show remaining requests if getting close to limit
- ✅ Disable submit button temporarily when rate limited

**DON'T**:
- ❌ Hide rate limit errors (confuses users)
- ❌ Retry automatically without user consent
- ❌ Show technical error codes without explanation

---

## Configuration Options

You can modify rate limits in `internal/http/ratelimit.go`:

### Change Request Limit

```go
func init() {
    limiter = &RateLimiter{
        visitors: make(map[string]*Visitor),
        limit:    200,  // Changed from 100 to 200
        window:   time.Hour,
    }
    // ...
}
```

### Change Time Window

```go
func init() {
    limiter = &RateLimiter{
        visitors: make(map[string]*Visitor),
        limit:    100,
        window:   30 * time.Minute,  // Changed from 1 hour to 30 minutes
    }
    // ...
}
```

### Different Limits for Different Endpoints

```go
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Determine limit based on method
        var limit int
        if r.Method == "GET" {
            limit = 200  // More lenient for reads
        } else {
            limit = 100  // Stricter for writes
        }
        
        // ... rest of logic
    })
}
```

---

## Architecture Details

### In-Memory Storage

**Structure**:
```go
type RateLimiter struct {
    visitors map[string]*Visitor  // IP -> visitor info
    mu       sync.RWMutex         // Thread-safe access
    limit    int                  // Max requests per window
    window   time.Duration        // Time window (1 hour)
}

type Visitor struct {
    count     int        // Current request count
    lastReset time.Time  // When counter was last reset
}
```

**Example state**:
```
visitors = {
    "1.2.3.4": {count: 45, lastReset: 2026-02-17 10:00:00},
    "5.6.7.8": {count: 12, lastReset: 2026-02-17 10:30:00},
    "9.10.11.12": {count: 100, lastReset: 2026-02-17 09:45:00},
}
```

### Automatic Cleanup

A background goroutine runs every 5 minutes to remove stale entries:

```go
// Runs in background
go limiter.cleanupVisitors()

func (rl *RateLimiter) cleanupVisitors() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        // Remove IPs not seen in 2+ hours
        for ip, visitor := range rl.visitors {
            if time.Since(visitor.lastReset) > 2 * time.Hour {
                delete(rl.visitors, ip)
            }
        }
    }
}
```

**Why this matters**:
- Prevents memory leak from tracking old IPs forever
- Keeps map size manageable
- Improves performance (smaller map = faster lookups)

---

## Production Considerations

### Current Implementation (Demo)

**Pros**:
- ✅ Simple (no external dependencies)
- ✅ Fast (in-memory lookups)
- ✅ Automatic cleanup
- ✅ Perfect for single-server demos

**Cons**:
- ⚠️ Resets if server restarts
- ⚠️ Doesn't work across multiple servers
- ⚠️ Limited to server's available memory

### Upgrading for Production

For high-traffic production systems with multiple servers, consider:

**Option 1: Redis**
```go
// Use Redis for shared state across servers
import "github.com/go-redis/redis/v8"

// All servers check same Redis instance
// Survives restarts
// Scales horizontally
```

**Option 2: Database-backed**
```sql
-- Store rate limit info in PostgreSQL
CREATE TABLE rate_limits (
    ip_address TEXT PRIMARY KEY,
    request_count INTEGER,
    window_start TIMESTAMP
);
```

**For this demo**: In-memory is perfect! ✅

---

## Monitoring

### Check Current Status

You can monitor rate limit status by checking response headers:

```bash
# Check your current limit status
curl -i http://localhost:8080/stats | grep X-RateLimit

# Output:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 73
# X-RateLimit-Reset: 1708212000
```

### Logs

Rate limiting doesn't log by default (to avoid spam), but you can add logging:

```go
if !allowed {
    log.Printf("Rate limit exceeded for IP: %s (tried: %d, limit: %d)", 
        ip, limiter.limit+1, limiter.limit)
    // ... return 429
}
```

---

## Common Questions

### Q: What happens if I restart the server?

**A**: Rate limit counters reset. Each IP starts fresh at 0/100.

This is acceptable for a demo. For production, use Redis for persistence.

---

### Q: Can users bypass this by using a VPN?

**A**: Yes, each unique IP gets its own limit. 

This is a limitation of IP-based rate limiting. For stricter control, implement API keys (tracks per key, not per IP).

---

### Q: Does this work with my frontend on Vercel and backend on Railway?

**A**: Yes! The backend correctly detects the client's real IP even through proxies via `X-Forwarded-For` header.

---

### Q: Why 100 requests per hour?

**A**: Conservative limit for demo purposes:
- Normal usage: 10-20 requests/hour
- Allows testing without hitting limit
- Prevents abuse/spam
- Can be increased for production

---

### Q: Does this affect the worker?

**A**: No, rate limiting only applies to HTTP API requests. The background worker is unaffected.

---

## Summary

✅ **Implemented**: IP-based rate limiting (100 req/hour)
✅ **Protected**: All API endpoints
✅ **User-friendly**: Clear error messages with retry times
✅ **Production-ready**: For single-server deployments
✅ **Scalable**: Easy to upgrade to Redis if needed
✅ **Tested**: Automated test script included

Your API is now protected from abuse while remaining accessible for legitimate users! 🚀
