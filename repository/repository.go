package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/needs"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/ggwhite/go-masker"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

var (
	psql                   = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	feedsLocationTableName = "feeds_location"
)

type PgxIface interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	SendBatch(context.Context, *pgx.Batch) pgx.BatchResults
	Close()
}

type Repository struct {
	pool *pgxpool.Pool
}

type GetLocationsQuery struct {
	SwLat, SwLng, NeLat, NeLng         float64
	Timestamp                          int64
	Reason, Channel                    string
	ExtraParams                        bool
	IsLocationVerified, IsNeedVerified string
}

type myQueryTracer struct {
	log *zap.SugaredLogger
}

func (tracer *myQueryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData) context.Context {
	tracer.log.Debugw("Executing command", "sql", data.SQL, "args", data.Args)

	return ctx
}

func (tracer *myQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func New() *Repository {
	dbUrl := os.Getenv("DB_CONN_STR")
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse config: %v\n", err)
		os.Exit(1)
	}

	config.MinConns = 5
	config.MaxConns = 10
	config.ConnConfig.Tracer = &myQueryTracer{
		log: log.Logger().Sugar(),
	}

	pool, err := pgxpool.NewWithConfig(
		context.Background(),
		config,
	)
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

func (repo *Repository) GetLocations(getLocationsQuery *GetLocationsQuery) ([]feeds.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*25)
	defer cancel()

	selectBuilder := psql.
		Select("id",
			"latitude",
			"longitude",
			"entry_id",
			"epoch",
			"reason",
			"channel",
			"is_location_verified",
			"is_need_verified",
			"needs").
		From(feedsLocationTableName)

	if getLocationsQuery.ExtraParams == true {
		selectBuilder = selectBuilder.Column("extra_parameters")
	}

	if getLocationsQuery.SwLat != 0.0 || getLocationsQuery.SwLng != 0.0 || getLocationsQuery.NeLat != 0.0 || getLocationsQuery.NeLng != 0.0 {
		selectBuilder = selectBuilder.Where(sq.GtOrEq{"southwest_lat": getLocationsQuery.SwLat, "southwest_lng": getLocationsQuery.SwLng}).
			Where(sq.LtOrEq{"northeast_lat": getLocationsQuery.NeLat, "northeast_lng": getLocationsQuery.NeLng})
	}

	if getLocationsQuery.Timestamp != 0 {
		if getLocationsQuery.Channel != "ahbap_location" {
			selectBuilder = selectBuilder.Where("epoch >= ?", getLocationsQuery.Timestamp)
		}
	}

	if getLocationsQuery.Reason != "" {
		splitted := strings.Split(getLocationsQuery.Reason, ",")
		splittedFormatted := make([]string, 0, len(splitted))
		for _, s := range splitted {
			splittedFormatted = append(splittedFormatted, "%"+s+"%")
		}
		selectBuilder = selectBuilder.Where("reason ILIKE ANY(?)", splittedFormatted)
	}

	if getLocationsQuery.Channel != "" {
		splitted := strings.Split(getLocationsQuery.Channel, ",")
		splittedFormatted := make([]string, 0, len(splitted))
		for _, s := range splitted {
			splittedFormatted = append(splittedFormatted, "%"+s+"%")
		}
		selectBuilder = selectBuilder.Where("channel ILIKE ANY(?)", splittedFormatted)
	}

	if getLocationsQuery.IsLocationVerified != "" {
		selectBuilder = selectBuilder.Where(sq.Eq{"is_location_verified": getLocationsQuery.IsLocationVerified})
	}

	if getLocationsQuery.IsNeedVerified != "" {
		selectBuilder = selectBuilder.Where(sq.Eq{"is_need_verified": getLocationsQuery.IsNeedVerified})
	}

	selectBuilder = selectBuilder.Where(sq.Eq{"is_deleted": false})

	newSql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not format query : %w", err)
	}

	query, err := repo.pool.Query(ctx, newSql, args...)
	if err != nil {
		return nil, fmt.Errorf("could not query locations: %w", err)
	}

	var results []feeds.Location

	for query.Next() {
		var result feeds.Location
		if getLocationsQuery.ExtraParams {
			err := query.Scan(&result.ID,
				&result.Latitude,
				&result.Longitude,
				&result.EntryID,
				&result.Epoch,
				&result.Reason,
				&result.Channel,
				&result.IsLocationVerified,
				&result.IsNeedVerified,
				&result.Needs,
				&result.ExtraParameters,
			)
			if err != nil {
				continue
			}

			result.Loc = []float64{result.Latitude, result.Longitude}

			if *result.Channel == "twitter" || *result.Channel == "discord" || *result.Channel == "babala" {
				result.ExtraParameters = maskFields(result.ExtraParameters)
			}
		} else {
			err := query.Scan(&result.ID,
				&result.Latitude,
				&result.Longitude,
				&result.EntryID,
				&result.Epoch,
				&result.Reason,
				&result.Channel,
				&result.IsLocationVerified,
				&result.IsNeedVerified,
				&result.Needs)
			if err != nil {
				continue
			}
			result.Loc = []float64{result.Latitude, result.Longitude}
		}

		results = append(results, result)
	}

	return results, nil
}

func maskFields(extraParams *string) *string {
	if extraParams == nil || *extraParams == "" {
		return nil
	}

	var jsonMap map[string]interface{}
	extraParamsStr := strings.ReplaceAll(*extraParams, " nan,", "'',")
	extraParamsStr = strings.ReplaceAll(extraParamsStr, " nan}", "''}")
	extraParamsStr = strings.ReplaceAll(extraParamsStr, "\\", "")

	if err := jsoniter.Unmarshal([]byte(strings.ReplaceAll(extraParamsStr, "'", "\"")), &jsonMap); err != nil {
		return nil
	}

	jsonMap["tel"] = masker.Telephone(fmt.Sprintf("%v", jsonMap["tel"]))
	jsonMap["telefon"] = masker.Telephone(fmt.Sprintf("%v", jsonMap["telefon"]))
	jsonMap["numara"] = masker.Telephone(fmt.Sprintf("%v", jsonMap["numara"]))
	jsonMap["isim-soyisim"] = masker.Name(fmt.Sprintf("%v", jsonMap["isim-soyisim"]))
	jsonMap["name_surname"] = masker.Name(fmt.Sprintf("%v", jsonMap["name_surname"]))
	jsonMap["name"] = masker.Name(fmt.Sprintf("%v", jsonMap["name"]))
	marshal, _ := jsoniter.Marshal(jsonMap)
	s := string(marshal)
	return &s
}

func (repo *Repository) GetFeed(id int64) (*feeds.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	rawSql, args, err := psql.Select(
		"fe.id",
		"full_text",
		"is_resolved",
		"fe.channel",
		"fe.timestamp",
		"fe.epoch",
		"fe.extra_parameters",
		"fl.formatted_address",
		"fl.reason",
		"fl.latitude",
		"fl.longitude").
		From("feeds_entry as fe").InnerJoin(feedsLocationTableName + " as fl on fl.entry_id = fe.id").
		Where(sq.Eq{"fe.id": id}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("could not prepare select feed query: %w", err)
	}

	row := repo.pool.QueryRow(ctx, rawSql, args...)

	var feed feeds.Feed
	if err := row.Scan(
		&feed.ID,
		&feed.FullText,
		&feed.IsResolved,
		&feed.Channel,
		&feed.Timestamp,
		&feed.Epoch,
		&feed.ExtraParameters,
		&feed.FormattedAddress,
		&feed.Reason,
		&feed.Lat,
		&feed.Lng); err != nil {
		return nil, fmt.Errorf("could not query feed with id : %w", err)
	}

	if feed.Channel == "twitter" || feed.Channel == "discord" || feed.Channel == "babala" {
		feed.ExtraParameters = maskFields(feed.ExtraParameters)
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

func (repo *Repository) CreateFeed(ctx context.Context, feed feeds.Feed, location feeds.Location) (error, int64) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error transaction begin stage %w", err), 0
	}
	defer tx.Rollback(ctx)

	entryID, err := repo.createFeedEntry(ctx, tx, feed)
	if err != nil {
		return err, 0
	}

	location.EntryID = entryID
	if _, err = repo.createFeedLocation(ctx, tx, location); err != nil {
		return err, 0
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error transaction commit stage %w", err), 0
	}

	return nil, entryID
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
	rawSql, args, err := psql.Insert(feedsLocationTableName).
		Columns(
			"id", "formatted_address",
			"latitude", "longitude",
			"northeast_lat", "northeast_lng",
			"southwest_lat", "southwest_lng",
			"entry_id",
			"epoch", "reason", "channel", "extra_parameters").
		Suffix("RETURNING \"id\"").
		Values(location.EntryID, location.FormattedAddress,
			location.Latitude, location.Longitude,
			location.NortheastLat, location.NortheastLng,
			location.SouthwestLat, location.SouthwestLng,
			location.EntryID,
			location.Epoch, location.Reason, location.Channel, location.ExtraParameters).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("could not prepare insert feeds location: %w", err)
	}

	var id int64

	if location.FormattedAddress != "" && location.Latitude != 0 && location.Longitude != 0 {
		err := tx.QueryRow(ctx, rawSql, args...).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("could not insert feeds location: %w", err)
		}
	}

	return id, nil
}

func (repo *Repository) UpdateLocationIntentAndNeeds(ctx context.Context, id int64, intents string, needs []feeds.NeedItem) error {
	updateBuilder := psql.Update(feedsLocationTableName).
		Set("reason", intents).
		Set("needs", needs).Where(sq.Eq{"entry_id": id})

	rawSql, args, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("could not prepare sql: %w", err)
	}

	if _, err := repo.pool.Exec(ctx, rawSql, args...); err != nil {
		return fmt.Errorf("could not update feeds location intent and needs: %w", err)
	}

	return nil
}

func (repo *Repository) DeleteFeedLocation(ctx context.Context, entryID int64) error {
	sql, args, err := psql.Update(feedsLocationTableName).Set("is_deleted", true).Where(sq.Eq{"entry_id": entryID}).ToSql()
	if err != nil {
		return fmt.Errorf("could not prepare soft delete query: %w", err)
	}
	_, err = repo.pool.Exec(ctx, sql, args...)
	return err
}

func (repo *Repository) UpdateFeedLocations(ctx context.Context, locations []feeds.FeedLocation) error {
	batch := &pgx.Batch{}
	for _, location := range locations {
		batch.Queue(
			"UPDATE "+feedsLocationTableName+" SET is_verified = true, latitude = $1, longitude = $2, formatted_address = $3 WHERE entry_id = $4;",
			location.Latitude, location.Longitude, location.Address, location.EntryID)
	}
	_, err := repo.pool.SendBatch(ctx, batch).Exec()
	return err
}
