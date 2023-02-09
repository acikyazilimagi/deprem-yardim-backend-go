package main

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/broker"
	"github.com/acikkaynak/backend-api-go/cache"
	"github.com/acikkaynak/backend-api-go/handler"
	"github.com/acikkaynak/backend-api-go/repository"
	_ "github.com/acikkaynak/backend-api-go/swagger"
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	app           *fiber.App
	repo          *repository.Repository
	kafkaProducer sarama.SyncProducer
}

func (a *Application) Register() {
	a.app.Get("/", handler.RedirectSwagger)
	a.app.Get("/healthcheck", handler.Healtcheck)
	a.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	a.app.Get("/monitor", monitor.New())
	a.app.Get("/feeds/areas", handler.GetFeedAreas(a.repo))
	a.app.Get("/feeds/:id/", handler.GetFeedById(a.repo))
	// We need to set up authentication for POST /events endpoint.
	a.app.Post("/events", handler.CreateEventHandler(a.kafkaProducer))
	route := a.app.Group("/swagger")
	route.Get("*", swagger.HandlerDefault)
}

// @title               IT Afet Yardım
// @version             1.0
// @description         Afet Harita API
// @host                127.0.0.1:80
// @BasePath            /
// @schemes             http https
func main() {
	repo := repository.New()
	defer repo.Close()
	cacheRepo := cache.NewRedisRepository()

	needsHandler := handler.NewNeedsHandler(repo)

	kafkaProducer, err := broker.NewProducer()
	if err != nil {
		log.Fatalf("failed to init kafka produder. err: %s", err)
	}

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

	app.Get("/needs", needsHandler.HandleList)
	app.Post("/needs", needsHandler.HandleCreate)

	application := &Application{app: app, repo: repo, kafkaProducer: kafkaProducer}
	application.Register()

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
