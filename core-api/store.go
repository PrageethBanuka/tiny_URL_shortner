package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errNotFound = errors.New("not found")


type store struct {
	db *pgxpool.Pool
}

func newStore(ctx context.Context, dsn string) (*store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &store{db: pool}, nil
}
func (s *store) close() {
	s.db.Close()
}

func (s *store) save(ctx context.Context, code, url string, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO links (code, url, expires_at) VALUES ($1, $2, $3)`,
		code, url, expiresAt,
	)
	return err
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