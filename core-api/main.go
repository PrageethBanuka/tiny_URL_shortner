package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	s := &store{links: make(map[string]link)}

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