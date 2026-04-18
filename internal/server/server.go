package server

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/klipitkas/machine-agent/internal/collector"
)

func New(addr string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /metadata", handleMetadata)
	mux.HandleFunc("GET /health", handleHealth)

	var handler http.Handler = mux
	if token := os.Getenv("TOKEN"); token != "" {
		handler = authMiddleware(token, mux)
	}
	handler = loggingMiddleware(handler)

	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			if subtle.ConstantTimeCompare([]byte(auth[7:]), []byte(token)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}
		if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("token")), []byte(token)) == 1 {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, r.RemoteAddr)
	})
}

func parseSections(r *http.Request) map[string]bool {
	include := r.URL.Query().Get("include")
	if include == "all" {
		sections := make(map[string]bool)
		for name := range collector.AllSections {
			sections[name] = true
		}
		return sections
	}
	if include != "" {
		sections := make(map[string]bool)
		for _, s := range strings.Split(include, ",") {
			s = strings.TrimSpace(s)
			if _, ok := collector.AllSections[s]; ok {
				sections[s] = true
			}
		}
		if len(sections) > 0 {
			return sections
		}
	}
	return collector.DefaultSections
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	sections := parseSections(r)
	info := collector.Collect(r.Context(), sections)

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(info)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
