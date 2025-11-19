package store

// PostgresStore is a placeholder for a PostgreSQL-backed storage implementation.
type PostgresStore struct{}

// NewPostgresStore creates a new PostgresStore.
// TODO: wire this up to a real PostgreSQL connection.
func NewPostgresStore() *PostgresStore {
	return &PostgresStore{}
}
