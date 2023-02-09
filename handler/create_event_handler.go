package handler

import (
	"fmt"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateEventHandler(producer sarama.SyncProducer) fiber.Handler {
	type request struct {
		ID        string         `json:"id"`
		Payload   string         `json:"payload"`
		Channel   string         `json:"channel"`
		Metadata  map[string]any `json:"metadata"`
		Timestamp int64          `json:"timestamp"`
	}

	return func(ctx *fiber.Ctx) error {
		var req request

		if err := ctx.BodyParser(&req); err != nil {
			return fmt.Errorf("failed to decode request. err: %w", err)
		}

		if len(req.ID) == 0 {
			req.ID = uuid.New().String()
		}

		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: req.Channel,
			Key:   sarama.StringEncoder(req.ID),
			Value: sarama.StringEncoder(req.Payload),
		})

		if err != nil {
			return fmt.Errorf("failed to send event. err: %w", err)
		}

		return ctx.SendStatus(http.StatusCreated)
	}
}
