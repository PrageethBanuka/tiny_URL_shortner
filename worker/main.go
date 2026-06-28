package main

import (
	"context" 
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp" 
)

func main() {
	tp, err := InitTracer("cleanup-worker")
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer func() {
		// Ensure traces flush before the cronjob exits
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()
	// -----------------------------------------------

	// Look for the environment variable OpenChoreo creates when you bind the dependency
	apiHost := os.Getenv("CORE_API_HOST")
	apiPort := os.Getenv("CORE_API_PORT")

	// Fallback to internal Kubernetes DNS if variables aren't set
	if apiHost == "" {
		apiHost = "core-api"
	}
	if apiPort == "" {
		apiPort = "8080" 
	}

	endpoint := fmt.Sprintf("http://%s:%s/internal/cleanup", apiHost, apiPort)
	log.Printf("CronJob started: Calling %s...", endpoint)

	// --- UPDATED: Wrap the HTTP Client Transport with OTel ---
	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   10 * time.Second,
	}

	// --- UPDATED: Use context to propagate traces ---
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
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