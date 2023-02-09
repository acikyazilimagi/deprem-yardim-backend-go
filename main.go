package main

import (
	"fmt"
	"github.com/acikkaynak/backend-api-go/handler"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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

	needsHandler := handler.NewNeedsHandler(repo)

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
		if c.Method() != http.MethodGet {
			// Don't cache write endpoints. We can maintain of list to exclude certain http methods later.
			// Since there will be an update in db, better to remove cache entries for this url
			err := cacheRepo.Delete(hashURL)
			if err != nil {
				fmt.Println(err)
			}
			return c.Next()
		}
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
		swLatStr := ctx.Query("sw_lat")
		swLngStr := ctx.Query("sw_lng")
		neLatStr := ctx.Query("ne_lat")
		neLngStr := ctx.Query("ne_lng")

		swLat, _ := strconv.ParseFloat(swLatStr, 64)
		swLng, _ := strconv.ParseFloat(swLngStr, 64)
		neLat, _ := strconv.ParseFloat(neLatStr, 64)
		neLng, _ := strconv.ParseFloat(neLngStr, 64)

		data, err := repo.GetLocations(swLat, swLng, neLat, neLng)
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

	app.Get("/needs", needsHandler.HandleList)
	app.Post("/needs", needsHandler.HandleCreate)

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
