package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	getLocationsQuery = "SELECT " +
		"fl.id, " +
		"formatted_address, " +
		"latitude, " +
		"longitude, " +
		"northeast_lat, " +
		"northeast_lng, " +
		"southwest_lat, " +
		"southwest_lng, " +
		"fe.full_text, " +
		"fe.timestamp," +
		"fe.extra_parameters " +
		"FROM feeds_location fl " +
		"inner join feeds_entry fe on fe.id = fl.entry_id " +
		"where southwest_lat >= %f " +
		"and southwest_lng >= %f " +
		"and northeast_lat <= %f " +
		"and northeast_lng <= %f"
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

func (repo *Repository) GetLocations(swLat, swLng, neLat, neLng float64) ([]feeds.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	query, err := repo.pool.Query(ctx, fmt.Sprintf(getLocationsQuery, swLat, swLng, neLat, neLng))
	if err != nil {
		return nil, fmt.Errorf("could not query locations: %w", err)
	}

	var results []feeds.Result

	for query.Next() {
		var result feeds.Result
		result.Loc = make([]float64, 2)

		err := query.Scan(&result.ID,
			&result.FormattedAddress,
			&result.Loc[0],
			&result.Loc[1],
			&result.ViewPort.Northeast.Lat,
			&result.ViewPort.Northeast.Lng,
			&result.ViewPort.Southwest.Lat,
			&result.ViewPort.Southwest.Lng,
			&result.Raw.FullText,
			&result.Raw.Timestamp,
			&result.Raw.ExtraParameters)
		if err != nil {
			continue
			//return nil, fmt.Errorf("could not scan locations: %w", err)
		}

		results = append(results, result)
	}

	return results, nil
}

func (repo *Repository) GetFeed(id int64) (*feeds.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	row := repo.pool.QueryRow(ctx, fmt.Sprintf("SELECT id, full_text, is_resolved, channel, extra_parameters FROM feeds_entry WHERE id=%d", id))

	var feed feeds.Feed
	if err := row.Scan(&feed.ID, &feed.FullText, &feed.IsResolved, &feed.Channel, &feed.ExtraParameters); err != nil {
		return nil, fmt.Errorf("could not query feed with id : %w", err)
	}

	return &feed, nil
}
