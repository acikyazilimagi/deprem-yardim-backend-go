package broker

import (
	"fmt"
	"log"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func Consume(group string, topic string, bootstrapServers string) {
	//HAS A HUGE LISTENING FOR LOOP BE CAREFUL
	
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		//	"bootstrap.servers": "host1:9092,host2:9092", be careful it is just a string
		"bootstrap.servers": bootstrapServers,
		"group.id":          group,
		"auto.offset.reset": "smallest"})

	if err != nil {
		log.Fatal(err)
	}
	
	err = consumer.Subscribe(topic, nil)

	if err != nil {
		log.Fatal(err)
	}
	
	for {
		ev := consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			fmt.Printf("reading with %s %s -> %s\n", group, topic, string(e.Value))
		case *kafka.Error:
			fmt.Printf("err %+V\n", err)
		default:
			fmt.Printf("why am i doing this i dont know ::\n",ev)
		}
	}
}
