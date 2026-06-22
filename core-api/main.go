package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	fmt.Println("core-api is listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
