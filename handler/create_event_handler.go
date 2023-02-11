package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type request struct {
	ID              string `json:"id"`
	RawText         string `json:"raw_text"`
	Channel         string `json:"channel"`
	ExtraParameters string `json:"extra_parameters"`
	Epoch           int64  `json:"epoch"`
}

// CreateEventHandler godoc
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

		if len(req.ID) == 0 {
			req.ID = uuid.New().String()
		}

		bytes, _ := json.Marshal(req)

		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: fmt.Sprintf("topic.raw.%s", req.Channel),
			Key:   sarama.StringEncoder(req.ID),
			Value: sarama.ByteEncoder(bytes),
		})

		if err != nil {
			return fmt.Errorf("failed to send event. err: %w", err)
		}

		return ctx.SendStatus(http.StatusCreated)
	}
}
