package progress

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ProgressStore abstracts database access for progress operations.
type ProgressStore interface {
	GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error)
	GetStudentProgressByTopic(ctx context.Context, arg queries.GetStudentProgressByTopicParams) ([]queries.GetStudentProgressByTopicRow, error)
	ListCompletedSessions(ctx context.Context, studentID pgtype.UUID) ([]queries.ListCompletedSessionsRow, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
}

// QueriesStore implements ProgressStore using sqlc-generated queries.
type QueriesStore struct {
	q *queries.Queries
}

// NewQueriesStore creates a new ProgressStore backed by sqlc queries.
func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

func (s *QueriesStore) GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error) {
	return s.q.GetStudentProgressBySubject(ctx, studentID)
}

func (s *QueriesStore) GetStudentProgressByTopic(ctx context.Context, arg queries.GetStudentProgressByTopicParams) ([]queries.GetStudentProgressByTopicRow, error) {
	return s.q.GetStudentProgressByTopic(ctx, arg)
}

func (s *QueriesStore) ListCompletedSessions(ctx context.Context, studentID pgtype.UUID) ([]queries.ListCompletedSessionsRow, error) {
	return s.q.ListCompletedSessions(ctx, studentID)
}

func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}
