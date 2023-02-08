package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tweet struct {
	ID         int
	FullText   string
	IsResolved bool
}

type Raw struct {
	Channel    string `json:"channel"`
	FullText   string `json:"full_text"`
	ID         int64  `json:"id"`
	IsResolved bool   `json:"is_resolved"`
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ViewPort struct {
	Northeast LatLng `json:"northeast"`
	Southwest LatLng `json:"southwest"`
}

type Result struct {
	FormattedAddress string    `json:"formatted_address"`
	ID               int64     `json:"id"`
	Loc              []float64 `json:"loc"`
	Raw              Raw       `json:"raw"`
	ViewPort         ViewPort  `json:"view_port"`
}

type Response struct {
	Count   int      `json:"count"`
	Results []Result `json:"results"`
}

var (
	pool *pgxpool.Pool
)

func getLocations(ctx context.Context, swLat, swLng, neLat, neLng float64) ([]Result, error) {
	query, err := pool.Query(context.Background(), fmt.Sprintf("SELECT id, formatted_address, latitude, longitude, northeast_lat, northeast_lng, southwest_lat, southwest_lng from feeds_location where southwest_lat >= %f and southwest_lng >= %f  and northeast_lat <= %f and northeast_lng <= %f", swLat, swLng, neLat, neLng))
	if err != nil {
		return nil, fmt.Errorf("could not query locations: %w", err)
	}

	var results []Result

	for query.Next() {
		var result Result
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

func main() {
	dbUrl := os.Getenv("DB_CONN_STR")
	var err error
	pool, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	app := fiber.New()
	app.Use(cors.New())

	feedsGroup := app.Group("/feeds")

	feedsGroup.Get("/areas", func(ctx *fiber.Ctx) error {
		swLatStr := ctx.Query("sw_lat")
		swLngStr := ctx.Query("sw_lng")
		neLatStr := ctx.Query("ne_lat")
		neLngStr := ctx.Query("ne_lng")

		swLat, _ := strconv.ParseFloat(swLatStr, 64)
		swLng, _ := strconv.ParseFloat(swLngStr, 64)
		neLat, _ := strconv.ParseFloat(neLatStr, 64)
		neLng, _ := strconv.ParseFloat(neLngStr, 64)

		data, err := getLocations(ctx.UserContext(), swLat, swLng, neLat, neLng)
		if err != nil {
			return ctx.JSON(err)
		}

		resp := Response{
			Count:   len(data),
			Results: data,
		}

		return ctx.JSON(resp)
	})

	app.Get("/healthcheck", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(":8000"); err != nil {
		panic("could not start app")
	}
}
