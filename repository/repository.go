package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/acikkaynak/backend-api-go/needs"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	getLocationsQuery = "SELECT " +
		"id, " +
		"latitude, " +
		"longitude, " +
		"entry_id, " +
		"timestamp, " +
		"epoch, " +
		"reason " +
		"FROM feeds_location " +
		"where southwest_lat >= %f " +
		"and southwest_lng >= %f " +
		"and northeast_lat <= %f " +
		"and northeast_lng <= %f " +
		"and epoch >= %d"
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

func (repo *Repository) GetLocations(swLat, swLng, neLat, neLng float64, timestamp int64, reason string) ([]feeds.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	q := fmt.Sprintf(getLocationsQuery, swLat, swLng, neLat, neLng, timestamp)
	if reason != "" {
		q = fmt.Sprintf("%s and reason = '%s'", q, reason)
	}
	query, err := repo.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("could not query locations: %w", err)
	}

	var results []feeds.Result

	for query.Next() {
		var result feeds.Result
		result.Loc = make([]float64, 2)

		err := query.Scan(&result.ID,
			&result.Loc[0],
			&result.Loc[1],
			&result.Entry_ID,
			&result.Timestamp,
			&result.Epoch,
			&result.Reason)
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
	row := repo.pool.QueryRow(ctx, fmt.Sprintf(
		"SELECT fe.id, full_text, is_resolved, fe.channel, fe.timestamp, fe.extra_parameters, fl.formatted_address, fl.reason "+
			"FROM feeds_entry fe, feeds_location fl "+
			"WHERE fe.id = fl.entry_id AND fe.id=%d", id))

	var feed feeds.Feed
	if err := row.Scan(&feed.ID, &feed.FullText, &feed.IsResolved, &feed.Channel, &feed.Timestamp, &feed.ExtraParameters, &feed.FormattedAddress, &feed.Reason); err != nil {
		return nil, fmt.Errorf("could not query feed with id : %w", err)
	}

	if feed.ExtraParameters != nil {
		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(*feed.ExtraParameters), &jsonMap)
		delete(jsonMap, "tel")
		delete(jsonMap, "name_surname")
		marshal, _ := json.Marshal(jsonMap)
		s := string(marshal)
		feed.ExtraParameters = &s
	}

	return &feed, nil
}

func (repo *Repository) GetNeeds(onlyNotResolved bool) ([]needs.Need, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	q := "SELECT n.id, n.description, n.is_resolved, n.timestamp, n.extra_parameters, n.formatted_address, n.latitude, n.longitude " +
		"FROM needs n"
	if onlyNotResolved {
		q = fmt.Sprintf("%s WHERE n.is_resolved=false", q)
	}
	query, err := repo.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("could not query needs: %w", err)
	}

	var results []needs.Need
	for query.Next() {
		var result needs.Need
		result.Loc = make([]float64, 2)

		err := query.Scan(&result.ID,
			&result.Description,
			&result.IsResolved,
			&result.Timestamp,
			&result.ExtraParameters,
			&result.FormattedAddress,
			&result.Loc[0],
			&result.Loc[1])
		if err != nil {
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

func (repo *Repository) CreateNeed(address, description string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	q := `INSERT INTO needs(address, description, timestamp, is_resolved, formatted_address, latitude, longitude) VALUES ($1::varchar, $2::varchar, $3::timestamp, $4::bool, $5::varchar, $6::int, $7::int) RETURNING id`

	var id int64
	err := repo.pool.QueryRow(ctx, q, address, description, time.Now(), false, "", 0, 0).Scan(&id)
	if err != nil {
		return id, fmt.Errorf("could not query needs: %w", err)
	}

	return id, nil
}
