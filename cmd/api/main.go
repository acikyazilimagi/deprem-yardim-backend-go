package main

import (
	"fmt"
	"github.com/acikkaynak/backend-api-go/search"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2/middleware/pprof"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/broker"
	"github.com/acikkaynak/backend-api-go/handler"
	"github.com/acikkaynak/backend-api-go/middleware/auth"
	"github.com/acikkaynak/backend-api-go/middleware/cache"
	"github.com/acikkaynak/backend-api-go/repository"
	_ "github.com/acikkaynak/backend-api-go/swagger"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Application struct {
	app           *fiber.App
	repo          *repository.Repository
	index         *search.LocationIndex
	kafkaProducer sarama.SyncProducer
}

func (a *Application) Register() {
	a.app.Get("/", handler.RedirectSwagger)
	a.app.Get("/healthcheck", handler.HealthCheck)
	a.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	a.app.Get("/monitor", monitor.New())
	a.app.Get("/feeds/areas", handler.GetFeedAreas(a.repo, a.index))
	a.app.Patch("/feeds/areas", handler.UpdateFeedLocationsHandler(a.repo))
	a.app.Get("/feeds/:id/", handler.GetFeedById(a.repo))
	a.app.Post("/events", handler.CreateEventHandler(a.kafkaProducer))
	a.app.Get("/caches/prune", handler.InvalidateCache())
	a.app.Get("/reasons", handler.GetReasonsHandler(a.repo))
	needsHandler := handler.NewNeedsHandler(a.repo)
	a.app.Get("/needs", needsHandler.HandleList)
	a.app.Post("/needs", needsHandler.HandleCreate)
	route := a.app.Group("/swagger")
	route.Get("*", swagger.HandlerDefault)
}

// @title						Afet Harita API
// @version					    1.0
// @description				    This is a sample swagger for Afet Harita
// @host						apigo.afetharita.com
// @BasePath					/
// @schemes					    https http
// @license.name				Apache License, Version 2.0 (the "License")
// @license.url				    https://github.com/acikkaynak/deprem-yardim-backend-go/blob/main/LICENSE
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						X-Api-Key
func main() {
	repo := repository.New()
	defer repo.Close()

	index := search.NewLocationIndex()

	kafkaProducer, err := broker.NewProducer()
	if err != nil {
		log.Println("failed to init kafka produder. err:", err)
	}

	app := fiber.New()
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))
	app.Use(cors.New())
	app.Use(recover.New())
	app.Use(auth.New())
	app.Use(pprof.New())
	app.Use(cache.New())

	application := &Application{app: app, repo: repo, index: index, kafkaProducer: kafkaProducer}
	application.Register()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
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
