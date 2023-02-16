package broker

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/acikkaynak/backend-api-go/pkg/logger"
)

func NewProducer() (sarama.SyncProducer, error) {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Logger().Panic("KAFKA_BROKERS env variable must be set")
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// We need to change below configs when kafka cluster was created
	// We don't know how we can connect to kafka
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer(brokers, config)
}
