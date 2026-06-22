package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	s := &store{links: make(map[string]link)}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/shorten", s.shortenHandler)
	mux.HandleFunc("/", s.resolveHandler)

	fmt.Println("core-api is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
