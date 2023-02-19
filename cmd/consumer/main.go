package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/acikkaynak/backend-api-go/consumer"

	"github.com/acikkaynak/backend-api-go/broker"
	"github.com/acikkaynak/backend-api-go/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	consumerGroupName = "feeds_location_consumer"
)

var (
	clientCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "go_consumer_metrics",
	}, []string{"topic", "timestamp"})
)

// Message will be handled in ConsumeClaim method.
func main() {
	http.HandleFunc("/healthcheck", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(":80", nil); err != nil {
			log.Logger().Error("server could not started or stopped", zap.Error(err))
		}
	}()

	client, err := broker.NewConsumerGroup(consumerGroupName)
	if err != nil {
		log.Logger().Panic(err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := consumer.NewConsumer(&client)
	c.Counter = clientCounter
	c.Start(ctx, &clientCounter)
	if err != nil {
		log.Logger().Panic(err.Error())
		return
	}

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
