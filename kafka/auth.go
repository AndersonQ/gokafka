package kafka

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"golang.org/x/oauth2"
)

// TODO
// returns a error channel for errors, the channel must be drained, otherwise it may lead to a memory leak
func NewRefreshTokenConsumer(
	tokenSource oauth2.TokenSource,
	consumer *kafka.Consumer,
	pollTimeout int) (func(), chan error) {

	errChan := make(chan error, 1)

	return func() {
		for {
			ev := consumer.Poll(pollTimeout)
			_, ok := ev.(kafka.OAuthBearerTokenRefresh)
			if !ok {
				continue
			}

			// TODO: delete debug
			log.Printf("refresh token event received: %#v", ev)
			token, err := tokenSource.Token()
			if err != nil {
				err := consumer.SetOAuthBearerTokenFailure(err.Error())
				if err != nil {
					// send the error in another goroutine so it won't block if the channel is not drained
					go func() { errChan <- fmt.Errorf("failure calling SetOAuthBearerTokenFailure: %w", err) }()
				}
				continue
			}

			err = consumer.SetOAuthBearerToken(kafka.OAuthBearerToken{
				TokenValue: token.AccessToken,
				Expiration: token.Expiry,
			})
			if err != nil {
				// send the error in another goroutine so it won't block if the channel is not drained
				go func() { errChan <- fmt.Errorf("error setting OAuthBearerToken: %w", err) }()
			}
		}
	}, errChan
}

// NewAuthenticatedConsumer takes a *github.com/confluentinc/confluent-kafka-go/kafka.ConfigMap,
// adds the needed configuration for OAuth 2 and returns a new kafka consumer
func NewAuthenticatedConsumer(cfg *kafka.ConfigMap) (*kafka.Consumer, error) {
	if err := cfg.SetKey("sasl.mechanism", "OAUTHBEARER"); err != nil {
		return nil, fmt.Errorf("could not set sasl.mechanism config: %w", err)
	}

	if err := cfg.SetKey("security.protocol", "SASL_SSL"); err != nil {
		return nil, fmt.Errorf("could not set sasl.mechanism config: %w", err)
	}

	return kafka.NewConsumer(cfg)
}
