package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
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

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "ok")
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

	s.mu.Lock()
	s.links[code] = link{URL: req.URL, ExpiresAt: expiresAt}
	s.mu.Unlock()

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

	s.mu.Lock()
	item, ok := s.links[code]
	if ok && time.Now().After(item.ExpiresAt) {
		delete(s.links, code)
		ok = false
	}
	s.mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, item.URL, http.StatusFound)
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
