package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crypto/rand"
	"math/big"
)

type shortenRequest struct {
	URL        string `json:"url"`
	TTLSeconds int    `json:"ttl_seconds,omitempty"`
}

type shortenResponse struct {
	Code      string `json:"code"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

func (s *store) healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Liveness only checks the process itself, not its dependencies —
	// that distinction is why this one stays a no-op 200, while readyz below doesn't.
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
}

func (s *store) readyzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.db.Ping(r.Context()); err != nil {
		http.Error(w, "database unreachable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ready")
}

func (s *store) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		http.Error(w, "url must start with http:// or https://", http.StatusBadRequest)
		return
	}

	ttl := 24 * time.Hour
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	code := generateCode(6)
	expiresAt := time.Now().Add(ttl)

	if err := s.save(r.Context(), code, req.URL, expiresAt); err != nil {
        // This will print the EXACT PostgreSQL rejection reason to your browser/curl
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{
		Code:      code,
		URL:       req.URL,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}

func (s *store) resolveHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" || strings.Contains(code, "/") {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url, err := s.resolve(r.Context(), code)
	if errors.Is(err, errNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (s *store) cleanupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	purged, err := s.cleanup(r.Context())
	if err != nil {
		http.Error(w, "cleanup failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"purged": purged})
}

func generateCode(length int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			result[i] = alphabet[i%len(alphabet)]
			continue
		}
		result[i] = alphabet[n.Int64()]
	}

	return string(result)
}