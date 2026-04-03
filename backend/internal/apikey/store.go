package apikey

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// QueriesStore wraps queries.Queries to implement APIKeyStore.
type QueriesStore struct {
	q *queries.Queries
}

// NewQueriesStore creates a new QueriesStore backed by sqlc-generated queries.
func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

// UpsertAPIKey stores or updates the encrypted API key for a parent.
func (s *QueriesStore) UpsertAPIKey(parentID pgtype.UUID, encKey []byte) (*APIKeyRecord, error) {
	row, err := s.q.UpsertAPIKey(context.Background(), queries.UpsertAPIKeyParams{
		ParentID:     parentID,
		EncryptedKey: encKey,
	})
	if err != nil {
		return nil, err
	}
	return &APIKeyRecord{
		EncryptedKey: row.EncryptedKey,
		CreatedAt:    timeFromPgtz(row.CreatedAt),
		UpdatedAt:    timeFromPgtz(row.UpdatedAt),
	}, nil
}

// GetAPIKeyByParentID retrieves the API key record for a parent.
// Returns ErrNotFound if no key exists.
func (s *QueriesStore) GetAPIKeyByParentID(parentID pgtype.UUID) (*APIKeyRecord, error) {
	row, err := s.q.GetAPIKeyByParentID(context.Background(), parentID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &APIKeyRecord{
		EncryptedKey: row.EncryptedKey,
		CreatedAt:    timeFromPgtz(row.CreatedAt),
		UpdatedAt:    timeFromPgtz(row.UpdatedAt),
	}, nil
}

// DeleteAPIKeyByParentID removes the API key for a parent.
func (s *QueriesStore) DeleteAPIKeyByParentID(parentID pgtype.UUID) error {
	return s.q.DeleteAPIKeyByParentID(context.Background(), parentID)
}

// timeFromPgtz converts pgtype.Timestamptz to time.Time.
func timeFromPgtz(t pgtype.Timestamptz) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

// isNotFoundError checks if the error indicates no rows found.
func isNotFoundError(err error) bool {
	return err != nil && err.Error() == "no rows in result set"
}
