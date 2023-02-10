package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

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
		"reason, " +
		"channel " +
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

func (repo *Repository) GetLocations(swLat, swLng, neLat, neLng float64, timestamp int64, reason, channel string) ([]feeds.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	q := fmt.Sprintf(getLocationsQuery, swLat, swLng, neLat, neLng, timestamp)
	if reason != "" {
		q = fmt.Sprintf("%s and reason = '%s'", q, reason)
	}
	if channel != "" {
		q = fmt.Sprintf("%s and channel = '%s'", q, channel)
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
			&result.Reason,
			&result.Channel)
		if err != nil {
			continue
			// return nil, fmt.Errorf("could not scan locations: %w", err)
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

func (repo *Repository) CreateFeed(ctx context.Context, feed feeds.Feed, location feeds.Location) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error transaction begin stage %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = repo.createFeedEntry(ctx, tx, feed); err != nil {
		return err
	}

	if _, err = repo.createFeedLocation(ctx, tx, location); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error transaction commit stage %w", err)
	}

	return nil
}

func (repo *Repository) createFeedEntry(ctx context.Context, tx pgx.Tx, feed feeds.Feed) (int64, error) {
	q := `INSERT INTO feeds_entry (
				full_text, is_resolved, channel, 
				extra_parameters, "timestamp", epoch,
				is_geolocated, reason
			)
			values (
				$1, $2, $3,
				$4, $5, $6,
				$7, $8
			) RETURNING id;`

	var id int64
	err := tx.QueryRow(ctx, q,
		feed.FullText, feed.IsResolved, feed.Channel,
		feed.ExtraParameters, feed.Timestamp, feed.Epoch,
		false, feed.Reason).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not insert feeds entry: %w", err)
	}

	return id, nil
}

func (repo *Repository) createFeedLocation(ctx context.Context, tx pgx.Tx, location feeds.Location) (int64, error) {
	q := `INSERT INTO feeds_location (
			formatted_address, 
			latitude, longitude, 
			northeast_lat, northeast_lng, 
			southwest_lat, southwest_lng, 
			entry_id, "timestamp", 
			epoch, reason, channel
		) values (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12
		) RETURNING id;`

	var id int64
	err := tx.QueryRow(ctx, q,
		location.FormattedAddress,
		location.Latitude, location.Longitude,
		location.NortheastLat, location.NortheastLng,
		location.SouthwestLat, location.SouthwestLng,
		location.EntryID, location.Timestamp,
		location.Epoch, location.Reason, location.Channel,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not insert feeds location: %w", err)
	}

	return id, nil
}

func (repo *Repository) UpdateLocationIntent(ctx context.Context, id int64, intents string) error {
	q := "UPDATE feeds_location SET reason = $1 WHERE id=$2;"

	_, err := repo.pool.Exec(ctx, q, intents, id)
	if err != nil {
		return fmt.Errorf("could not update feeds location intent: %w", err)
	}

	return nil
}
