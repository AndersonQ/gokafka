package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Handler interface {
	Handle(context.Context, *kafka.Message)
}

type HandlerFunc func(ctx context.Context, msg *kafka.Message)

func (h HandlerFunc) Handle(ctx context.Context, msg *kafka.Message) {
	h(ctx, msg)
}

// timeout milliseconds
func RunConsumer(kafkaConsumer *kafka.Consumer, handler Handler) {
	go func() {
		for {
			log.Print("waiting message")
			msg, err := kafkaConsumer.ReadMessage(-1)
			if err != nil {
				log.Print(fmt.Sprintf("[ERROR] failed to read message: %v", err))
				continue
			}

			handler.Handle(context.Background(), msg)
		}
	}()
}
