package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog"

	"github.com/AndersonQ/gokafka/config"
	myKafka "github.com/AndersonQ/gokafka/kafka"
	"github.com/AndersonQ/gokafka/oauth"
)

func handleEvent(ctx context.Context, msg *kafka.Message) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msgf("message.TopicPartition: %s, message.Key: %s", msg.TopicPartition)
}

func main() {
	// catch the signals as soon as possible
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt) // a.k.a ctrl+C
	signal.Notify(signalChan, os.Kill)      // a.k.a kill

	cfg, err := config.Parse()
	if err != nil {
		panic("could not parse environment variables: " + err.Error())
	}

	logger := cfg.Logger()
	tokenW := oauth.NewTokenSource(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL, 5*time.Second, http.Client{
		Timeout: cfg.RequestTimeout,
	})

	consumer, err := myKafka.NewAuthenticatedConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       cfg.KafkaServer,
		"group.id":                cfg.KafkaGroupID,
		"enable.auto.commit":      "false",
		"auto.commit.interval.ms": "1000",
		"auto.offset.reset":       "earliest",
	})

	refreshTokenConsumer, errChan := myKafka.NewRefreshTokenConsumer(tokenW, consumer, 1000)
	logger.Info().Msg("Starting refreshTokenConsumer")
	go refreshTokenConsumer()
	go func() {
		for err := range errChan {
			logger.Error().Err(err).Msg("failures on refreshTokenConsumer")
		}
	}()

	timeoutMiddleware := myKafka.NewTimeoutMiddleware(cfg.RequestTimeout)
	logMiddleware := myKafka.NewLoggerMiddleware(logger)

	h := timeoutMiddleware(myKafka.HandlerFunc(handleEvent))
	h = logMiddleware(h)
	h = myKafka.TrackingIDMiddleware(h)
	h = myKafka.LogRequestMiddleware(h)

	logger.Info().Msg("Consuming messages...")
	myKafka.RunConsumer(consumer, h)

	<-signalChan
	logger.Info().Msg("done!")
}
