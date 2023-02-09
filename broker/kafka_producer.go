package broker

import (
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

// Go üzerinde hiç kullanmadım lakin internet yardımı ile ortaya bir şeyler çıktı
// Kod okunabilirlik için kısaltılmıştır ve error gibi fonksiyonları utils adında bir dosya da tutmak okunabilirliği arttırmak ile beraber projeyi hızlandırır.

config := sarama.NewConfig()
config.Version = sarama.V2_5_0  

brokers := []string{"localhost:9092"}  // Default port

client, err := sarama.NewClient(brokers, config) // Hata olmadığı sürece dosyalar arası client objesi sorunsuz çalışabilir ve tekrar statement oluşturmayız: "new"
if err != nil { panic(err) }

defer func() { 
	     if err := client.Close(); 
	      err != nil { panic(err) } 
	     }()


// Producer aşağıda ki gibidir ve consumer aracılığı ile okunabilir.
// Başka bir dosyaya alınabilir örn: router'a dayatılabilir.

producer, err := sarama.NewSyncProducerFromClient(client)

if err != nil { panic(err) }
defer func() { if err := producer.Close(); err != nil { panic(err) } }()

message := &sarama.ProducerMessage{
    Topic: "test-topic",
    Value: sarama.StringEncoder("test message"),
}

partition, offset, err := producer.SendMessage(message)
if err != nil { panic(err) }

fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
