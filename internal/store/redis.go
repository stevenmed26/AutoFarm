package store

// RedisStore is a placeholder for a Redis-backed cache implementation.
type RedisStore struct{}

// NewRedisStore creates a new RedisStore.
// TODO: wire this up to a real Redis client.
func NewRedisStore() *RedisStore {
	return &RedisStore{}
}
