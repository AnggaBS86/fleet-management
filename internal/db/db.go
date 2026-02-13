package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	store := &Store{pool: pool}
	if err := store.initSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) initSchema(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS vehicle_locations (
	id BIGSERIAL PRIMARY KEY,
	vehicle_id TEXT NOT NULL,
	latitude DOUBLE PRECISION NOT NULL,
	longitude DOUBLE PRECISION NOT NULL,
	timestamp BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_vehicle_locations_vehicle_time
ON vehicle_locations (vehicle_id, timestamp DESC);
`
	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Store) InsertLocation(ctx context.Context, vehicleID string, lat, lon float64, ts int64) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp) VALUES ($1, $2, $3, $4)`,
		vehicleID, lat, lon, ts,
	)
	return err
}

type Location struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func (s *Store) GetLatest(ctx context.Context, vehicleID string) (*Location, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT vehicle_id, latitude, longitude, timestamp
		 FROM vehicle_locations
		 WHERE vehicle_id = $1
		 ORDER BY timestamp DESC
		 LIMIT 1`, vehicleID)

	var loc Location
	if err := row.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (s *Store) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]Location, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT vehicle_id, latitude, longitude, timestamp
		 FROM vehicle_locations
		 WHERE vehicle_id = $1 AND timestamp BETWEEN $2 AND $3
		 ORDER BY timestamp ASC`, vehicleID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Location
	for rows.Next() {
		var loc Location
		if err := rows.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, err
		}
		result = append(result, loc)
	}
	return result, rows.Err()
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}
