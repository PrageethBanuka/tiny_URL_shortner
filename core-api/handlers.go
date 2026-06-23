package main

import (
    "encoding/json"
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
    code := strings.TrimPrefix(r.URL.Path, "/")
    
    // Explicitly ignore internal API routes so they don't trigger a database lookup
    if code == "" || code == "shorten" || code == "healthz" || code == "readyz" || code == "metrics" || strings.Contains(code, "/") {
        http.NotFound(w, r)
        return
    }

    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Query Postgres directly for the original URL
    var targetURL string
    query := `SELECT url FROM links WHERE short_id = $1 AND expires_at > NOW()`
    err := s.db.QueryRow(r.Context(), query, code).Scan(&targetURL)
    
    if err != nil {
        http.NotFound(w, r)
        return
    }

    // Execute the HTTP 302 Redirect
    http.Redirect(w, r, targetURL, http.StatusFound)
}

func (s *store) cleanupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Execute the Worker cleanup query against Postgres
    query := `DELETE FROM links WHERE expires_at < NOW()`
    result, err := s.db.Exec(r.Context(), query)
    if err != nil {
        http.Error(w, "cleanup failed", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    // Pgx Exec returns a tag containing the number of rows affected
    _ = json.NewEncoder(w).Encode(map[string]int64{"purged": result.RowsAffected()})
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