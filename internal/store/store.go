package store

import (
	"context"
	"encoding/json"

	"github.com/gioboa/go-gke-alloydb/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{
		Pool:    pool,
		Queries: db.New(pool),
	}, nil
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil || s.Pool == nil {
		return nil
	}
	return s.Pool.Ping(ctx)
}

func (s *Store) ListRegions(ctx context.Context, limit int) ([]map[string]any, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT row_to_json(r)
		FROM (
			SELECT *
			FROM regions
			ORDER BY region_id
			LIMIT $1
		) AS r
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regions := make([]map[string]any, 0, limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}

		var region map[string]any
		if err := json.Unmarshal(raw, &region); err != nil {
			return nil, err
		}
		regions = append(regions, region)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return regions, nil
}

func (s *Store) Close() {
	if s == nil || s.Pool == nil {
		return
	}
	s.Pool.Close()
}
