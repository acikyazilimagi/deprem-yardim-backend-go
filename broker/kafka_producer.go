package broker

import (
	"log"
	"os"
	"strings"
	//"github.com/Shopify/sarama"
)
kafkaBrokers := os.Getenv("KAFKA_BROKERS")


func NewProducer() (sarama.SyncProducer, error) {
	
	if kafkaBrokers == "" {
		log.Panic("KAFKA_BROKERS env variable must be set")
	}
	brokers := strings.Split(kafkaBrokers, ",")

	// We need to change below configs when kafka cluster was created
	// We don't know how we can connect to kafka
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer(brokers, config)
}
