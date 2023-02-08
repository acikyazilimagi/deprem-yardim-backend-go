package main

import (
	"fmt"
	"strconv"

	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	repo := repository.New()
	defer repo.Close()

	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover2.New())
	app.Use(monitor.New())

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

		data, err := repo.GetLocations(ctx.UserContext(), swLat, swLng, neLat, neLng)
		if err != nil {
			return ctx.JSON(err)
		}

		resp := feeds.Response{
			Count:   len(data),
			Results: data,
		}

		return ctx.JSON(resp)
	})

	app.Get("/healthcheck", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(":80"); err != nil {
		panic(fmt.Sprintf("app error: %s", err.Error()))
	}
}
