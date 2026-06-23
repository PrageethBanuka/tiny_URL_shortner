package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func dsnFromEnv() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || password == "" {
		log.Fatal("DB_HOST and DB_PASSWORD are required")
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)
}

// Middleware to allow cross-origin requests from your frontend
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // In production, change "*" to your exact frontend domain
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

        // Browsers send a preflight OPTIONS request before the actual request
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}


func main() {
	dsn := dsnFromEnv()
	if dsn == "" {
		log.Fatal("DATABASE_URL is required, e.g. postgres://postgres:devpassword@localhost:5432/core_api")
	}

	ctx := context.Background()
	s, err := newStore(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer s.close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", withMetrics("/healthz", s.healthzHandler))
	mux.HandleFunc("/readyz", withMetrics("/readyz", s.readyzHandler))
	mux.HandleFunc("/shorten", withMetrics("/shorten", s.shortenHandler))
	mux.HandleFunc("/internal/cleanup", withMetrics("/internal/cleanup", s.cleanupHandler))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", withMetrics("/resolve", s.resolveHandler))

	fmt.Println("core-api is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux)))
}