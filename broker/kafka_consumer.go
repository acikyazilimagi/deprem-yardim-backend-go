package broker

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

func NewConsumerGroup(group string) (sarama.ConsumerGroup, error) {
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	// We need to change below configs when kafka cluster was created
	// We don't know how we can connect to kafka
	config := sarama.NewConfig()

	return sarama.NewConsumerGroup(brokers, group, config)
}
