package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

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
		Addr:    addr,
		Handler: handler,
	}
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") && auth[7:] == token {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Query().Get("token") == token {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
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
	info, err := collector.Collect(r.Context(), sections)
	if err != nil {
		http.Error(w, `{"error":"failed to collect metadata"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(info)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
