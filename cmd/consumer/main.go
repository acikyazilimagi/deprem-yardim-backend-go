package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/broker"
	"github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	consumerGroupName             = "feeds_location_consumer"
	intentConsumerGroupName       = "feeds_intent_consumer"
	addressResolvedTopicName      = "topic.feeds.location"
	intentResolvedTopicName       = "topic.feeds.intent"
	AWS_TASK_METADATA_URL_ENV_VAR = "ECS_CONTAINER_METADATA_URI_V4"
)

var (
	clientCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_consumer_metrics",
	}, []string{"topic", "timestamp", "task_id"})

	taskID string
)

// Message will be handled in ConsumeClaim method.
func main() {
	http.HandleFunc("/healthcheck", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(":80", nil); err != nil {
			fmt.Fprintf(os.Stderr, "server could not started or stopped: %s", err)
		}
	}()

	client, err := broker.NewConsumerGroup(consumerGroupName)
	if err != nil {
		log.Logger().Panic(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := NewConsumer()
	go func() {
		for {
			if err := client.Consume(ctx, []string{intentResolvedTopicName, addressResolvedTopicName}, consumer); err != nil {
				log.Logger().Panic("Error from consumer:", zap.Error(err))
			}
			// check if context was cancelled, signaling that the addressResolvedConsumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Logger().Info("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	healthy := true
	for healthy {
		select {
		case <-ctx.Done():
			log.Logger().Info("terminating: context cancelled")
			healthy = false
		case <-sigterm:
			log.Logger().Info("terminating: via signal")
			healthy = false
		}
	}

	cancel()
	if err = client.Close(); err != nil {
		log.Logger().Panic("Error closing client:", zap.Error(err))
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready    chan bool
	repo     *repository.Repository
	producer sarama.SyncProducer
}

func NewConsumer() *Consumer {
	producer, err := broker.NewProducer()
	if err != nil {
		log.Logger().Panic("failed to init kafka producer. err:", zap.Error(err))
	}
	return &Consumer{
		ready:    make(chan bool),
		repo:     repository.New(),
		producer: producer,
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
			clientCounter.With(prometheus.Labels{
				"topic":     message.Topic,
				"timestamp": fmt.Sprintf("%d", message.Timestamp.Unix()),
				"task_id":   taskID,
			}).Inc()
			if message.Topic == intentResolvedTopicName {
				consumer.intentResolveHandle(message, session)
			}
			if message.Topic == addressResolvedTopicName {
				consumer.addressResolveHandle(message, session)
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

type DuplicationRequest struct {
	Address string   `json:"address"`
	Intents []string `json:"reasons"`
	Needs   []string `json:"needs"`
}

type DuplicationResponse struct {
	IsDuplicate bool `json:"is_duplicate"`
}
