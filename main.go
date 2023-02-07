package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Tweet struct {
	ID                   int
	CreatedAt            time.Duration
	Verified             bool
	UserID               int64
	ScreenName           int64
	UserAccountCreatedAt time.Time
	FullText             string
	Name                 string
	Media                string
	Hashtags             string
	TweetId              string
	AddressResolved      bool
}

func main() {
	app := fiber.New()

	locationGroup := app.Group("/locations")

	locationGroup.Get("/areas", func(ctx *fiber.Ctx) error {
		return ctx.JSON(nil)
	})

	locationGroup.Get("/areas/count", func(ctx *fiber.Ctx) error {
		return ctx.JSON(nil)
	})

	app.Get("/cities", func(ctx *fiber.Ctx) error {
		return ctx.JSON(nil)
	})

	app.Get("/healthcheck", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(":8000"); err != nil {
		panic("could not start app")
	}
}
