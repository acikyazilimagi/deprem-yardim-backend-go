package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/acikkaynak/backend-api-go/cache"
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	repo := repository.New()
	defer repo.Close()
	cacheRepo := cache.NewRedisRepository()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover2.New())
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/healthcheck" ||
			c.Path() == "/metrics" ||
			c.Path() == "/monitor" {
			return c.Next()
		}

		reqURI := c.OriginalURL()
		hashURL := uuid.NewSHA1(uuid.NameSpaceOID, []byte(reqURI)).String()
		cacheData := cacheRepo.Get(hashURL)

		if cacheData == nil {
			c.Next()
			cacheRepo.SetKey(hashURL, c.Response().Body(), 0)
			return nil
		}

		return c.JSON(cacheData)
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	app.Get("/monitor", monitor.New())

	app.Get("/feeds/areas", func(ctx *fiber.Ctx) error {
		layout := "2006-01-02T15:04:05.000-03:00"

		swLatStr := ctx.Query("sw_lat")
		swLngStr := ctx.Query("sw_lng")
		neLatStr := ctx.Query("ne_lat")
		neLngStr := ctx.Query("ne_lng")
		timeStampStr := ctx.Query("time_stamp")

		swLat, _ := strconv.ParseFloat(swLatStr, 64)
		swLng, _ := strconv.ParseFloat(swLngStr, 64)
		neLat, _ := strconv.ParseFloat(neLatStr, 64)
		neLng, _ := strconv.ParseFloat(neLngStr, 64)
		dataTime, err := time.Parse(layout, timeStampStr)
		if err != nil {
			fmt.Sprintf("time parse error %s", err)
		}

		data, err := repo.GetLocations(swLat, swLng, neLat, neLng, dataTime)
		if err != nil {
			return ctx.JSON(err)
		}

		resp := &feeds.Response{
			Count:   len(data),
			Results: data,
		}

		return ctx.JSON(resp)
	})

	app.Get("/feeds/:id/", func(ctx *fiber.Ctx) error {
		feedIDStr := ctx.Params("id")

		feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
		if err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		feed, err := repo.GetFeed(feedID)
		if err != nil {
			return ctx.JSON(err)
		}

		return ctx.JSON(feed)
	})

	app.Get("/healthcheck", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		_ = <-c
		fmt.Println("application gracefully shutting down..")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":80"); err != nil {
		panic(fmt.Sprintf("app error: %s", err.Error()))
	}
}
