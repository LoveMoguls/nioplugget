package content

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ContentStore abstracts database access for content browsing.
type ContentStore interface {
	ListSubjects(ctx context.Context) ([]queries.ListSubjectsRow, error)
	GetSubjectBySlug(ctx context.Context, slug string) (queries.GetSubjectBySlugRow, error)
	ListTopicsBySubjectID(ctx context.Context, subjectID pgtype.UUID) ([]queries.ListTopicsBySubjectIDRow, error)
	GetTopicBySlug(ctx context.Context, params queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error)
	ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error)
}

// QueriesStore implements ContentStore using sqlc-generated queries.
type QueriesStore struct {
	q *queries.Queries
}

// NewQueriesStore creates a new QueriesStore.
func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

func (s *QueriesStore) ListSubjects(ctx context.Context) ([]queries.ListSubjectsRow, error) {
	return s.q.ListSubjects(ctx)
}

func (s *QueriesStore) GetSubjectBySlug(ctx context.Context, slug string) (queries.GetSubjectBySlugRow, error) {
	return s.q.GetSubjectBySlug(ctx, slug)
}

func (s *QueriesStore) ListTopicsBySubjectID(ctx context.Context, subjectID pgtype.UUID) ([]queries.ListTopicsBySubjectIDRow, error) {
	return s.q.ListTopicsBySubjectID(ctx, subjectID)
}

func (s *QueriesStore) GetTopicBySlug(ctx context.Context, params queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error) {
	return s.q.GetTopicBySlug(ctx, params)
}

func (s *QueriesStore) ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error) {
	return s.q.ListExercisesByTopicID(ctx, topicID)
}
