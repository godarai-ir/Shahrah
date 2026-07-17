package platformcredentials

import (
	"context"
	"encoding/json"
)

const invalidationChannel = "platform-provider-credentials:v1:invalidation"

type CoordinationClient interface {
	Publish(ctx context.Context, channel, message string) error
	Subscribe(ctx context.Context, channel string) (<-chan string, error)
}

type RedisInvalidationBus struct{ client CoordinationClient }

func NewRedisInvalidationBus(client CoordinationClient) *RedisInvalidationBus {
	if client == nil {
		return nil
	}
	return &RedisInvalidationBus{client: client}
}

func (b *RedisInvalidationBus) PublishPlatformCredentialInvalidation(ctx context.Context, event InvalidationEvent) error {
	encoded, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, invalidationChannel, string(encoded))
}

// Consume reloads outside the request path. Invalid payloads are ignored and
// the durable 30-second reconciler repairs missed or malformed messages.
func (b *RedisInvalidationBus) Consume(ctx context.Context, manager *Manager) error {
	messages, err := b.client.Subscribe(ctx, invalidationChannel)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case payload, ok := <-messages:
			if !ok {
				return nil
			}
			var event InvalidationEvent
			if json.Unmarshal([]byte(payload), &event) != nil || event.ConnectionID == "" || event.ProviderSlug == "" || event.Environment == "" {
				continue
			}
			manager.Invalidate(event.ProviderSlug, event.Environment)
			_ = manager.Reload(ctx)
		}
	}
}
