package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/feeds"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type FeedMessage struct {
	ID               int64     `json:"id,omitempty"`
	FullText         string    `json:"raw_text"`
	IsResolved       bool      `json:"is_resolved"`
	Channel          string    `json:"channel,omitempty"`
	Timestamp        time.Time `json:"timestamp,omitempty"`
	Epoch            int64     `json:"epoch"`
	ExtraParameters  *string   `json:"extra_parameters,omitempty"`
	FormattedAddress string    `json:"formatted_address,omitempty"`
	Reason           *string   `json:"reason,omitempty"`
}

type ConsumeMessagePayload struct {
	Location feeds.Location `json:"location"`
	Feed     FeedMessage    `json:"feed"`
}

func (consumer *Consumer) addressResolveHandle(message *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) {
	var messagePayload ConsumeMessagePayload
	if err := jsoniter.Unmarshal(message.Value, &messagePayload); err != nil {
		log.Logger().Error("addressResolveHandle deserialization error", zap.String("payload", string(message.Value)), zap.Error(err))
		session.MarkMessage(message, "")
		session.Commit()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	messagePayload.Feed.Timestamp = time.Now()

	f := feeds.Feed{
		ID:               messagePayload.Feed.ID,
		FullText:         messagePayload.Feed.FullText,
		IsResolved:       messagePayload.Feed.IsResolved,
		Channel:          messagePayload.Feed.Channel,
		Timestamp:        messagePayload.Feed.Timestamp,
		Epoch:            messagePayload.Feed.Epoch,
		ExtraParameters:  messagePayload.Feed.ExtraParameters,
		FormattedAddress: messagePayload.Feed.FormattedAddress,
		Reason:           messagePayload.Feed.Reason,
	}

	err, entryID := consumer.repo.CreateFeed(ctx, f, messagePayload.Location)
	if err != nil {
		log.Logger().Error("error inserting feed entry and location", zap.String("payload", string(message.Value)), zap.Error(err))
		return
	}

	intentPayloadByte, err := jsoniter.Marshal(IntentMessagePayload{
		FeedID:          entryID,
		FullText:        messagePayload.Feed.FullText,
		ResolvedAddress: messagePayload.Location.FormattedAddress,
		Location:        messagePayload.Location,
	})

	_, _, err = consumer.producer.SendMessage(&sarama.ProducerMessage{
		Topic: IntentResolvedTopicName,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", entryID)),
		Value: sarama.ByteEncoder(intentPayloadByte),
	})
	if err != nil {
		log.Logger().Error("error producing intent", zap.Int64("feedID", messagePayload.Feed.ID), zap.Error(err))
		session.MarkMessage(message, "")
		session.Commit()
		return
	}

	session.MarkMessage(message, "")
	session.Commit()
}
