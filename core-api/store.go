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


func (s *store) cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	purged := 0
	now := time.Now()
	for code, l := range s.links {
		if now.After(l.ExpiresAt) {
			delete(s.links, code)
			purged++
		}
	}
	return purged
}