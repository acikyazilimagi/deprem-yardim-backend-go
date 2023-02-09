package broker

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

func NewProducer() (sarama.SyncProducer, error) {
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	// We need to change below configs when kafka cluster was created
	// We don't know how we can connect to kafka
	config := sarama.NewConfig()
	// Return success is required for sync producer.
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer(brokers, config)
}
