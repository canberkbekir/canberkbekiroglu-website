package analytics

import (
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

// metricsWriter wraps http.ResponseWriter to capture the status code.
type metricsWriter struct {
	http.ResponseWriter
	status int
}

func (mw *metricsWriter) WriteHeader(code int) {
	mw.status = code
	mw.ResponseWriter.WriteHeader(code)
}

// Tracker returns chi middleware that records page visits and request metrics.
// It skips static files for visitor tracking but records metrics for all non-static requests.
func (db *DB) Tracker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		start := time.Now()

		mw := &metricsWriter{ResponseWriter: w, status: http.StatusOK}

		if shouldTrack(urlPath) {
			ip := extractIP(r)
			go func() {
				if err := db.RecordVisit(ip, urlPath); err != nil {
					log.Printf("Analytics tracking error: %v", err)
				}
			}()
		}

		next.ServeHTTP(mw, r)

		// Record request metrics for non-static paths.
		if !strings.HasPrefix(urlPath, "/static/") {
			duration := time.Since(start).Milliseconds()
			go func() {
				if err := db.RecordMetrics(urlPath, r.Method, mw.status, int(duration)); err != nil {
					log.Printf("Metrics recording error: %v", err)
				}
			}()
		}
	})
}

// shouldTrack returns false for paths that should not be recorded.
func shouldTrack(urlPath string) bool {
	if strings.HasPrefix(urlPath, "/static/") {
		return false
	}
	if strings.HasPrefix(urlPath, "/api/") {
		return false
	}
	if strings.HasPrefix(urlPath, "/admin") {
		return false
	}
	// Skip paths with file extensions (favicon.ico, robots.txt, etc.).
	if ext := path.Ext(urlPath); ext != "" {
		return false
	}
	return true
}

// TrustProxy controls whether X-Forwarded-For and X-Real-IP headers are trusted.
// Set to true only when running behind a trusted reverse proxy (nginx, Cloudflare, etc.).
var TrustProxy = false

// extractIP pulls the client IP, only trusting proxy headers when TrustProxy is enabled.
func extractIP(r *http.Request) string {
	if TrustProxy {
		// X-Forwarded-For: client, proxy1, proxy2
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.SplitN(xff, ",", 2)
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// RemoteAddr is "ip:port".
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
