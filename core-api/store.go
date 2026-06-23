package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errNotFound = errors.New("not found")


type store struct {
	db *pgxpool.Pool
}

func newStore(ctx context.Context, dsn string) (*store, error) {
    // 1. Establish the connection pool
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %w", err)
    }
    
    // 2. Verify the connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("unable to ping database: %w", err)
    }

    // 3. Auto-Migrate: Ensure the links table exists
    // (Adjust the column names if your shortenHandler uses different ones)
    schema := `
    DROP TABLE IF EXISTS links;
    CREATE TABLE links (
        short_id VARCHAR(20) PRIMARY KEY,
        url TEXT NOT NULL,
        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`
    
    _, err = pool.Exec(ctx, schema)
    if err != nil {
        return nil, fmt.Errorf("failed to create database schema: %w", err)
    }

    return &store{db: pool}, nil
}
func (s *store) close() {
	s.db.Close()
}

func (s *store) save(ctx context.Context, code, targetURL string, expiresAt time.Time) error {
    query := `
        INSERT INTO links (short_id, url, expires_at) 
        VALUES ($1, $2, $3)
    `
    
    // Execute the query using the exact Postgres driver syntax
    _, err := s.db.Exec(ctx, query, code, targetURL, expiresAt)
    if err != nil {
        // Return the actual database error so we can read it!
        return fmt.Errorf("db insert failed: %w", err) 
    }
    
    return nil
}

func (s *store) resolve(ctx context.Context, code string) (string, error) {
	var url string
	var expiresAt time.Time

	err := s.db.QueryRow(ctx,
		`SELECT url, expires_at FROM links WHERE code = $1`, code,
	).Scan(&url, &expiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", errNotFound
	}
	if err != nil {
		return "", err
	}

	if time.Now().After(expiresAt) {
		_, _ = s.db.Exec(ctx, `DELETE FROM links WHERE code = $1`, code)
		return "", errNotFound
	}

	return url, nil
}

func (s *store) cleanup(ctx context.Context) (int, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM links WHERE expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}


// type link struct {
// 	URL       string
// 	ExpiresAt time.Time
// }

// type store struct {
// 	mu    sync.Mutex
// 	links map[string]link
// }


// func (s *store) cleanup() int {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	purged := 0
// 	now := time.Now()
// 	for code, l := range s.links {
// 		if now.After(l.ExpiresAt) {
// 			delete(s.links, code)
// 			purged++
// 		}
// 	}
// 	return purged
// }