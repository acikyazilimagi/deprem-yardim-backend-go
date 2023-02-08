package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New() *Repository {
	dbUrl := os.Getenv("DB_CONN_STR")
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return &Repository{
		pool: pool,
	}
}

func (repo *Repository) Close() {
	repo.pool.Close()
}

func (repo *Repository) GetLocations(ctx context.Context, swLat, swLng, neLat, neLng float64) ([]feeds.Result, error) {
	query, err := repo.pool.Query(context.Background(), fmt.Sprintf("SELECT id, formatted_address, latitude, longitude, northeast_lat, northeast_lng, southwest_lat, southwest_lng from feeds_location where southwest_lat >= %f and southwest_lng >= %f  and northeast_lat <= %f and northeast_lng <= %f", swLat, swLng, neLat, neLng))
	if err != nil {
		return nil, fmt.Errorf("could not query locations: %w", err)
	}

	var results []feeds.Result

	for query.Next() {
		var result feeds.Result
		result.Loc = make([]float64, 2)
		var id int64

		err := query.Scan(&id,
			&result.FormattedAddress,
			&result.Loc[0],
			&result.Loc[1],
			&result.ViewPort.Northeast.Lat,
			&result.ViewPort.Northeast.Lng,
			&result.ViewPort.Southwest.Lat,
			&result.ViewPort.Southwest.Lng)
		if err != nil {
			continue
			//return nil, fmt.Errorf("could not scan locations: %w", err)
		}

		result.ID = id
		results = append(results, result)
	}

	return results, nil
}

func (repo *Repository) GetFeed(ctx context.Context, id int64) (*feeds.Feed, error) {
	row := repo.pool.QueryRow(context.Background(), fmt.Sprintf("SELECT id, full_text, is_resolved, channel FROM feeds_entry WHERE id=%d", id))

	var feed feeds.Feed
	if err := row.Scan(&feed.ID, &feed.FullText, &feed.IsResolved, &feed.Channel); err != nil {
		return nil, fmt.Errorf("could not query feed with id : %w", err)
	}

	return &feed, nil
}
