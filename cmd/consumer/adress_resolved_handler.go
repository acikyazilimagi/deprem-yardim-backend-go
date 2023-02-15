package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/feeds"
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
	if err := json.Unmarshal(message.Value, &messagePayload); err != nil {
		fmt.Fprintf(os.Stderr, "deserialization error message %s error %s", string(message.Value), err.Error())
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
		fmt.Fprintf(os.Stderr, "error inserting feed entry and location message %#v error %s rawMessage %s", messagePayload, err.Error(), string(message.Value))
		return
	}

	intentPayloadByte, err := json.Marshal(IntentMessagePayload{
		FeedID:          entryID,
		FullText:        messagePayload.Feed.FullText,
		ResolvedAddress: messagePayload.Location.FormattedAddress,
	})

	_, _, err = consumer.producer.SendMessage(&sarama.ProducerMessage{
		Topic: intentResolvedTopicName,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", entryID)),
		Value: sarama.ByteEncoder(intentPayloadByte),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error producing intent feedID %d error %s", messagePayload.Feed.ID, err.Error())
		session.MarkMessage(message, "")
		session.Commit()
		return
	}

	session.MarkMessage(message, "")
	session.Commit()
}
