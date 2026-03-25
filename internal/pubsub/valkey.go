package pubsub

import (
	"context"
	"sync"

	"github.com/valkey-io/valkey-go"
)

type RedisPubSub struct {
	val valkey.Client

	mu sync.Mutex
	cancels map[string]context.CancelFunc
}

func NewRedisPubSub(rdb valkey.Client) *RedisPubSub {
	return &RedisPubSub{
		val: rdb,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (p *RedisPubSub) Publish(ctx context.Context, topic string, msg string) {

	p.val.Do(
		ctx,
		p.val.B().
			Publish().
			Channel(topic).
			Message(msg).
			Build(),
	)
}

func (p *RedisPubSub) Subscribe(topic string) <-chan string {

	ch := make(chan string, 10)

	ctx, cancel := context.WithCancel(context.Background())

	p.mu.Lock()
	p.cancels[topic] = cancel
	p.mu.Unlock()

	go func() {

		defer close(ch)

		p.val.Receive(
			ctx,
			p.val.B().
				Subscribe().
				Channel(topic).
				Build(),

			func(msg valkey.PubSubMessage) {

				select {
				case ch <- msg.Message:
				default:
				}

			},
		)

	}()

	return ch
}

func (p *RedisPubSub) Unsubscribe(topic string, ch <-chan string) {

	p.mu.Lock()

	if cancel, ok := p.cancels[topic]; ok {
		cancel()
		delete(p.cancels, topic)
	}

	p.mu.Unlock()
}