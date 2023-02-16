package broker

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
)

func NewConsumerGroup(group string) (sarama.ConsumerGroup, error) {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Logger().Panic("KAFKA_BROKERS env variable must be set")
	}
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	// We need to change below configs when kafka cluster was created
	// We don't know how we can connect to kafka
	config := sarama.NewConfig()

	return sarama.NewConsumerGroup(brokers, group, config)
}
