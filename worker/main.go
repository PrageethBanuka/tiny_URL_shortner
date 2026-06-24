package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Look for the environment variable OpenChoreo creates when you bind the dependency
	apiHost := os.Getenv("CORE_API_HOST")
	apiPort := os.Getenv("CORE_API_PORT")

	// Fallback to internal Kubernetes DNS if variables aren't set
	if apiHost == "" {
		apiHost = "core-api"
	}
	if apiPort == "" {
		apiPort = "8080" // Assuming your core-api runs on 8080
	}

	endpoint := fmt.Sprintf("http://%s:%s/internal/cleanup", apiHost, apiPort)
	log.Printf("CronJob started: Calling %s...", endpoint)

	// Set a timeout so the worker doesn't hang forever if the API is slow
	client := &http.Client{Timeout: 10 * time.Second}

	// Assuming your /internal/cleanup endpoint uses POST. 
	// (Change to http.MethodGet if your handler expects a GET request).
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error calling core-api cleanup: %v\n", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Fatalf("Cleanup failed with status %d: %s\n", resp.StatusCode, string(body))
	}

	fmt.Printf("Cleanup request successful: %s\n", string(body))
	log.Println("CronJob complete. Shutting down gracefully.")
}