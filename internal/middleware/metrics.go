package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
)

// Metrics collects per-request Prometheus-style metrics and exposes them via
// the /metrics endpoint registered in cmd.go.
func Metrics(r *ghttp.Request) {
	atomic.AddInt64(&metricsInFlight, 1)
	start := time.Now()
	r.Middleware.Next()
	atomic.AddInt64(&metricsInFlight, -1)
	latency := time.Since(start).Seconds()
	status := r.Response.Status
	if status == 0 {
		status = 200
	}
	path := normalizeMetricPath(r.URL.Path)
	method := r.Method
	label := fmt.Sprintf(`method="%s",path="%s",status="%d"`, method, path, status)
	metricsMu.Lock()
	if metricsCounters == nil {
		metricsCounters = map[string]*metricEntry{}
	}
	entry, ok := metricsCounters[label]
	if !ok {
		entry = &metricEntry{Method: method, Path: path, Status: status}
		metricsCounters[label] = entry
	}
	entry.Count++
	entry.TotalLatency += latency
	metricsMu.Unlock()
}

// metricEntry holds accumulated observations for one label combination.
type metricEntry struct {
	Method       string
	Path         string
	Status       int
	Count        int64
	TotalLatency float64
}

var (
	metricsInFlight int64
	metricsMu       sync.Mutex
	metricsCounters map[string]*metricEntry
)

// RenderMetrics produces a Prometheus text-format exposition.
func RenderMetrics() string {
	var b strings.Builder
	b.WriteString("# HELP http_requests_total Total number of HTTP requests.\n")
	b.WriteString("# TYPE http_requests_total counter\n")
	metricsMu.Lock()
	entries := make([]*metricEntry, 0, len(metricsCounters))
	for _, e := range metricsCounters {
		entries = append(entries, e)
	}
	metricsMu.Unlock()
	for _, e := range entries {
		b.WriteString(fmt.Sprintf(`http_requests_total{method="%s",path="%s",status="%d"} %d`+"\n",
			e.Method, e.Path, e.Status, e.Count))
	}
	b.WriteString("\n# HELP http_request_duration_seconds_total Cumulative request duration in seconds.\n")
	b.WriteString("# TYPE http_request_duration_seconds counter\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf(`http_request_duration_seconds_total{method="%s",path="%s",status="%d"} %.6f`+"\n",
			e.Method, e.Path, e.Status, e.TotalLatency))
	}
	inFlight := atomic.LoadInt64(&metricsInFlight)
	b.WriteString("\n# HELP http_requests_in_flight Current number of in-flight requests.\n")
	b.WriteString("# TYPE http_requests_in_flight gauge\n")
	b.WriteString(fmt.Sprintf(`http_requests_in_flight %d`+"\n", inFlight))
	return b.String()
}

// normalizeMetricPath collapses high-cardinality path segments (UUIDs, slugs)
// into placeholders to keep metric cardinality bounded.
func normalizeMetricPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := range parts {
		if isUUIDLike(parts[i]) && i > 0 {
			parts[i] = ":id"
		}
	}
	normalized := "/" + strings.Join(parts, "/")
	return normalized
}

func isUUIDLike(s string) bool {
	if len(s) < 32 {
		return false
	}
	hasDash := false
	for _, c := range s {
		if c == '-' {
			hasDash = true
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return hasDash
}

func init() {
	metricsCounters = map[string]*metricEntry{}
}

// MetricsHandler returns a ghttp handler that serves /metrics.
func MetricsHandler(r *ghttp.Request) {
	r.Response.Header().Set("Content-Type", "text/plain; version=0.0.4")
	r.Response.Status = http.StatusOK
	r.Response.Write(RenderMetrics())
}
