package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

// RateLimit implements a token-bucket rate limiter keyed by IP + UserID.
// It is applied to /api/v1/* routes and returns 429 Too Many Requests when
// the bucket is exhausted.
func RateLimit(r *ghttp.Request) {
	cfg := rateLimitConfig(r.GetCtx())
	if !cfg.Enabled {
		r.Middleware.Next()
		return
	}
	key := rateLimitKey(r)
	bucket := acquireBucket(key, cfg.RPS, cfg.Burst)
	if !bucket.allow() {
		r.Response.Header().Set("Retry-After", "1")
		r.Response.Header().Set("X-RateLimit-Limit", itoa(cfg.Burst))
		r.Response.Header().Set("X-RateLimit-Remaining", "0")
		writeError(r, http.StatusTooManyRequests, "RATE_LIMITED", nil)
		return
	}
	r.Response.Header().Set("X-RateLimit-Limit", itoa(cfg.Burst))
	r.Response.Header().Set("X-RateLimit-Remaining", itoa(bucket.tokenCount()))
	r.Middleware.Next()
}

type rateLimitCfg struct {
	Enabled bool
	RPS     float64
	Burst   int
}

func rateLimitConfig(ctx context.Context) rateLimitCfg {
	return rateLimitCfg{
		Enabled: service.Config().GetBool(ctx, "rateLimit.enabled", false),
		RPS:     service.Config().GetFloat(ctx, "rateLimit.rps", 20.0),
		Burst:   service.Config().GetInt(ctx, "rateLimit.burst", 40),
	}
}

func rateLimitKey(r *ghttp.Request) string {
	identity, _ := service.AuthIdentityFromCtx(r.GetCtx())
	if identity != nil && identity.UserID != "" {
		return "uid:" + identity.UserID
	}
	return "ip:" + r.GetClientIp()
}

// ---- Token bucket implementation ----

type tokenBucket struct {
	mu       sync.Mutex
	rate     float64
	capacity int
	count    float64
	last     time.Time
}

var (
	bucketsMu sync.Mutex
	buckets   = map[string]*tokenBucket{}
)

func acquireBucket(key string, rps float64, burst int) *tokenBucket {
	bucketsMu.Lock()
	defer bucketsMu.Unlock()
	if b, ok := buckets[key]; ok {
		b.mu.Lock()
		b.rate = rps
		if burst > b.capacity {
			b.capacity = burst
		}
		b.mu.Unlock()
		return b
	}
	b := &tokenBucket{rate: rps, capacity: burst, count: float64(burst), last: time.Now()}
	buckets[key] = b
	return b
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.count += elapsed * b.rate
	if b.count > float64(b.capacity) {
		b.count = float64(b.capacity)
	}
	b.last = now
	if b.count < 1 {
		return false
	}
	b.count--
	return true
}

func (b *tokenBucket) tokenCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return int(b.count)
}

func itoa(n int) string {
	if n <= 0 {
		return "0"
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
