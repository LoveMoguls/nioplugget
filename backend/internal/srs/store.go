package srs

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// SRSStore abstracts database access for spaced repetition operations.
type SRSStore interface {
	UpsertReviewSchedule(ctx context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error)
	ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error)
	GetReviewSchedule(ctx context.Context, arg queries.GetReviewScheduleParams) (queries.ReviewSchedule, error)
}

// QueriesStore implements SRSStore using sqlc-generated queries.
type QueriesStore struct {
	q *queries.Queries
}

// NewQueriesStore creates a new SRS QueriesStore.
func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

func (s *QueriesStore) UpsertReviewSchedule(ctx context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error) {
	return s.q.UpsertReviewSchedule(ctx, arg)
}

func (s *QueriesStore) ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return s.q.ListDueReviews(ctx, studentID)
}

func (s *QueriesStore) GetReviewSchedule(ctx context.Context, arg queries.GetReviewScheduleParams) (queries.ReviewSchedule, error) {
	return s.q.GetReviewSchedule(ctx, arg)
}
