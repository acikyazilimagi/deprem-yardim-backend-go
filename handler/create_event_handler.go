package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	Feeds []RawFeed `json:"feeds"`
}

type RawFeed struct {
	ID              string `json:"id"`
	RawText         string `json:"raw_text"`
	Channel         string `json:"channel"`
	ExtraParameters string `json:"extra_parameters"`
	Epoch           int64  `json:"epoch"`
}

// createEvent godoc
// @Summary            Create Event areas with request body
// @Tags               Event
// @Accept             json
// @Produce            json
// @Success            200 {object} nil
// @Param              body body request true "RequestBody"
// @Router             /events [POST]
func CreateEventHandler(producer sarama.SyncProducer) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var req request

		if err := ctx.BodyParser(&req); err != nil {
			return fmt.Errorf("failed to decode request. err: %w", err)
		}

		for _, f := range req.Feeds {
			f.ID = uuid.New().String()
			bytes, _ := json.Marshal(f)

			_, _, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: fmt.Sprintf("topic.raw.%s", f.Channel),
				Key:   sarama.StringEncoder(f.ID),
				Value: sarama.ByteEncoder(bytes),
			})

			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to send event. err: %s", err.Error())
			}
		}

		return ctx.SendStatus(http.StatusCreated)
	}
}
