package pubsub

import (
	"context"
)

type PubSub interface {
    Publish(ctx context.Context, topic string, msg string)
    Subscribe(topic string) <-chan string
    Unsubscribe(topic string, ch <-chan string)
}