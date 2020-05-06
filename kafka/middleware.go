package kafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/AndersonQ/gokafka/constants"
)

// Now provides the current time, defaults to time.Now. Change if you want to control the time
var Now = time.Now

type key struct{}

var ctxKey key

func IDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ctxKey).(string); ok {
		return id
	}

	return ""
}

func ContextWithID(parent context.Context, id string) context.Context {
	return context.WithValue(parent, ctxKey, id)
}

func NewTimeoutMiddleware(timeout time.Duration) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *kafka.Message) {
			ctxx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			h.Handle(ctxx, msg)
		})
	}
}

func NewLoggerMiddleware(logger zerolog.Logger) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *kafka.Message) {
			ctxx := logger.WithContext(ctx)

			h.Handle(ctxx, msg)
		})
	}
}

// LogRequestMiddleware assumes there is a zerolog.Logger in the context
func LogRequestMiddleware(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, msg *kafka.Message) {
		logger := zerolog.Ctx(ctx)

		startedAt := Now()
		defer func() {
			l := logger.With().
				Dur(constants.FieldRequestDuration, startedAt.Sub(Now())).
				Str(constants.FieldEvent, constants.EventRequestFinished).Logger()
			l.Info().Msg("request finished")
		}()

		h.Handle(ctx, msg)
	})
}

// TrackingIDMiddleware assumes there is a zerolog.Logger in the context
func TrackingIDMiddleware(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, msg *kafka.Message) {
		logger := zerolog.Ctx(ctx)

		trackingID := getHeader(constants.HeaderTrackingID, msg.Headers)
		if trackingID == "" {
			id, err := uuid.NewUUID()
			if err != nil {
				logger.Panic().Err(err).Msg("could not generate trackingID")
			}
			trackingID = id.String()
		}

		ctx = ContextWithID(ctx, trackingID)

		logger.With().Str(constants.FieldTrackingID, trackingID)
		h.Handle(ctx, msg)
	})
}

func getHeader(key string, hs []kafka.Header) string {
	for _, h := range hs {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
