package app

import (
	"context"
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	app           *fiber.App
	repo          *repository.Repository
	kafkaProducer sarama.SyncProducer
}

func NewPoolConnection() *pgxpool.Pool {
	dbUrl := os.Getenv("DB_CONN_STR")
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return pool
}
