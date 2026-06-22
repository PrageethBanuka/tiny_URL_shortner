package main

import (
	"sync"
	"time"
)

type link struct {
	URL       string
	ExpiresAt time.Time
}

type store struct {
	mu    sync.Mutex
	links map[string]link
}
