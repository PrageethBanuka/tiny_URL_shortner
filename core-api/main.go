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
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)
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
	log.Fatal(http.ListenAndServe(":8080", mux))
}