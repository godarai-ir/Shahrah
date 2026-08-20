// Package store contains tenant-scoped metadata and vector persistence for managed RAG.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ErrLeaseLost indicates that another worker or a cancellation owns the job now.
var ErrLeaseLost = errors.New("job lease lost")

// Store owns independent metadata and vector database pools.
type Store struct {
	metadata        *pgxpool.Pool
	vectors         *pgxpool.Pool
	redis           *redis.Client
	conversationKey []byte
}

// New connects and verifies both managed RAG databases.
func New(ctx context.Context, metadataURL, vectorURL, redisURL string, conversationKey []byte) (*Store, error) {
	if metadataURL == "" || vectorURL == "" || redisURL == "" || len(conversationKey) < 32 {
		return nil, errors.New("metadata, vector, Redis, and conversation encryption configuration are required")
	}

	metadata, connectErr := pgxpool.New(ctx, metadataURL)
	if connectErr != nil {
		return nil, fmt.Errorf("connect metadata database: %w", connectErr)
	}

	if err := metadata.Ping(ctx); err != nil {
		metadata.Close()

		return nil, fmt.Errorf("ping metadata database: %w", err)
	}

	vectors, vectorConnectErr := pgxpool.New(ctx, vectorURL)
	if vectorConnectErr != nil {
		metadata.Close()

		return nil, fmt.Errorf("connect vector database: %w", vectorConnectErr)
	}

	if err := vectors.Ping(ctx); err != nil {
		vectors.Close()
		metadata.Close()

		return nil, fmt.Errorf("ping vector database: %w", err)
	}

	redisOptions, parseErr := redis.ParseURL(redisURL)
	if parseErr != nil {
		vectors.Close()
		metadata.Close()

		return nil, fmt.Errorf("parse Redis URL: %w", parseErr)
	}

	redisClient := redis.NewClient(redisOptions)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		closeRedis(redisClient)
		vectors.Close()
		metadata.Close()

		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	key := append([]byte(nil), conversationKey...)

	return &Store{metadata: metadata, vectors: vectors, redis: redisClient, conversationKey: key}, nil
}

// Close releases both database pools.
func (s *Store) Close() {
	if s == nil {
		return
	}

	if s.vectors != nil {
		s.vectors.Close()
	}

	if s.metadata != nil {
		s.metadata.Close()
	}

	if s.redis != nil {
		closeRedis(s.redis)
	}
}

// Ready verifies that both data stores are reachable.
func (s *Store) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.metadata.Ping(ctx); err != nil {
		return fmt.Errorf("metadata database: %w", err)
	}

	if err := s.vectors.Ping(ctx); err != nil {
		return fmt.Errorf("vector database: %w", err)
	}

	if err := s.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}

	return nil
}

func closeRedis(client *redis.Client) {
	if err := client.Close(); err != nil {
		// The caller is already shutting down or handling a connection failure.
		return
	}
}
