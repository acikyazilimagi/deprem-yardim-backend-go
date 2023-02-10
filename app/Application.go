package app

import (
	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/gofiber/fiber/v2"
)

type Application struct {
	app           *fiber.App
	repo          *repository.Repository
	kafkaProducer sarama.SyncProducer
}
