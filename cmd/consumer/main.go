package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/broker"
	"github.com/acikkaynak/backend-api-go/feeds"
	"github.com/acikkaynak/backend-api-go/repository"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	consumerGroupName = "feeds_location_consumer"
	topicName         = "topic.feeds.location"
)

// Message will be handled in ConsumeClaim method.
func main() {
	client, err := broker.NewConsumerGroup(consumerGroupName)
	if err != nil {
		log.Panic(err.Error())
		return
	}

	consumer := NewConsumer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if err := client.Consume(ctx, []string{topicName}, consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	healthy := true
	for healthy {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			healthy = false
		case <-sigterm:
			log.Println("terminating: via signal")
			healthy = false
		}
	}

	cancel()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready      chan bool
	repository *repository.Repository
}

func NewConsumer() *Consumer {
	return &Consumer{
		ready:      make(chan bool),
		repository: repository.New(),
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			var messagePayload ConsumeMessagePayload
			if err := json.Unmarshal(message.Value, &messagePayload); err != nil {
				fmt.Fprintf(os.Stderr, "deserialization error message %s error %s", string(message.Value), err.Error())
				session.MarkMessage(message, "")
				session.Commit()
				continue
			}

			err := consumer.Insert(messagePayload)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error inserting feed entry and location message %#v error %s rawMessage %s", messagePayload, err.Error(), string(message.Value))
				continue
			}

			session.MarkMessage(message, "")
			session.Commit()
		case <-session.Context().Done():
			return nil
		}
	}
}

func (consumer *Consumer) Insert(messagePayload ConsumeMessagePayload) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messagePayload.Feed.Timestamp = time.Now()

	return consumer.repository.CreateFeed(ctx, messagePayload.Feed, messagePayload.Location)
}

type ConsumeMessagePayload struct {
	Location feeds.Location `json:"location"`
	Feed     feeds.Feed     `json:"feed"`
}
