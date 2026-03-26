package store

import (
	"context"
	"errors"

	"github.com/gioboa/go-postgresql/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

type Region struct {
	ID   int64  `json:"region_id"`
	Name string `json:"region_name"`
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

func (s *Store) ListRegions(ctx context.Context, limit int) ([]Region, error) {
	rows, err := s.Queries.ListRegions(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	regions := make([]Region, 0, len(rows))
	for _, row := range rows {
		regions = append(regions, toRegion(row))
	}
	return regions, nil
}

func (s *Store) GetRegion(ctx context.Context, id int64) (Region, error) {
	region, err := s.Queries.GetRegion(ctx, id)
	if err != nil {
		return Region{}, err
	}
	return toRegion(region), nil
}

func (s *Store) CreateRegion(ctx context.Context, region Region) (Region, error) {
	created, err := s.Queries.CreateRegion(ctx, db.CreateRegionParams{
		RegionID:   region.ID,
		RegionName: textValue(region.Name),
	})
	if err != nil {
		return Region{}, err
	}
	return toRegion(created), nil
}

func (s *Store) UpdateRegion(ctx context.Context, id int64, name string) (Region, error) {
	region, err := s.Queries.UpdateRegion(ctx, db.UpdateRegionParams{
		RegionID:   id,
		RegionName: textValue(name),
	})
	if err != nil {
		return Region{}, err
	}
	return toRegion(region), nil
}

func (s *Store) DeleteRegion(ctx context.Context, id int64) (Region, error) {
	region, err := s.Queries.DeleteRegion(ctx, id)
	if err != nil {
		return Region{}, err
	}
	return toRegion(region), nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toRegion(region db.Region) Region {
	return Region{
		ID:   region.RegionID,
		Name: region.RegionName.String,
	}
}

func textValue(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func (s *Store) Close() {
	if s == nil || s.Pool == nil {
		return
	}
	s.Pool.Close()
}
