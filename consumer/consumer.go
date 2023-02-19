package consumer

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/acikkaynak/backend-api-go/broker"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/acikkaynak/backend-api-go/repository"
	"github.com/acikkaynak/backend-api-go/search"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	AddressResolvedTopicName = "topic.feeds.location"
	IntentResolvedTopicName  = "topic.feeds.intent"
)

type Consumer struct {
	Ready    chan bool
	repo     *repository.Repository
	index    *search.LocationIndex
	producer sarama.SyncProducer
	consumer *sarama.ConsumerGroup
	Counter  *prometheus.CounterVec
}

func NewConsumer(consumer *sarama.ConsumerGroup) *Consumer {
	producer, err := broker.NewProducer()
	if err != nil {
		log.Logger().Panic("failed to init kafka producer. err:", zap.Error(err))
	}
	return &Consumer{
		Ready:    make(chan bool),
		repo:     repository.New(),
		index:    search.NewLocationIndex(),
		producer: producer,
		consumer: consumer,
	}
}

func (consumer *Consumer) Start(ctx context.Context, i **prometheus.CounterVec) {
	go func() {
		for {
			if err := (*consumer.consumer).Consume(ctx, []string{IntentResolvedTopicName, AddressResolvedTopicName}, consumer); err != nil {
				log.Logger().Panic("Error from c:", zap.Error(err))
			}
			// check if context was cancelled, signaling that the addressResolvedConsumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()
	<-consumer.Ready
	log.Logger().Info("Sarama consumer up and running!...")
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the c as Ready
	close(consumer.Ready)
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
			consumer.Counter.With(prometheus.Labels{
				"topic":     message.Topic,
				"timestamp": fmt.Sprintf("%d", message.Timestamp.Unix()),
			}).Inc()
			if message.Topic == IntentResolvedTopicName {
				consumer.intentResolveHandle(message, session)
			}
			if message.Topic == AddressResolvedTopicName {
				consumer.addressResolveHandle(message, session)
			}
		case <-session.Context().Done():
			return nil
		}
	}
}
