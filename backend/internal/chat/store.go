package chat

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ChatStore abstracts database access for chat operations.
type ChatStore interface {
	CreateSession(ctx context.Context, arg queries.CreateSessionParams) (queries.Session, error)
	GetSessionByID(ctx context.Context, id pgtype.UUID) (queries.Session, error)
	EndSession(ctx context.Context, arg queries.EndSessionParams) (queries.Session, error)
	CreateMessage(ctx context.Context, arg queries.CreateMessageParams) (queries.Message, error)
	ListMessagesBySessionID(ctx context.Context, sessionID pgtype.UUID) ([]queries.Message, error)
	GetExerciseByID(ctx context.Context, id pgtype.UUID) (queries.GetExerciseByIDRow, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
	GetAPIKeyByParentID(ctx context.Context, parentID pgtype.UUID) (queries.ApiKey, error)
}

// QueriesStore implements ChatStore using sqlc-generated queries.
type QueriesStore struct {
	q *queries.Queries
}

// NewQueriesStore creates a new QueriesStore.
func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

func (s *QueriesStore) CreateSession(ctx context.Context, arg queries.CreateSessionParams) (queries.Session, error) {
	return s.q.CreateSession(ctx, arg)
}

func (s *QueriesStore) GetSessionByID(ctx context.Context, id pgtype.UUID) (queries.Session, error) {
	return s.q.GetSessionByID(ctx, id)
}

func (s *QueriesStore) EndSession(ctx context.Context, arg queries.EndSessionParams) (queries.Session, error) {
	return s.q.EndSession(ctx, arg)
}

func (s *QueriesStore) CreateMessage(ctx context.Context, arg queries.CreateMessageParams) (queries.Message, error) {
	return s.q.CreateMessage(ctx, arg)
}

func (s *QueriesStore) ListMessagesBySessionID(ctx context.Context, sessionID pgtype.UUID) ([]queries.Message, error) {
	return s.q.ListMessagesBySessionID(ctx, sessionID)
}

func (s *QueriesStore) GetExerciseByID(ctx context.Context, id pgtype.UUID) (queries.GetExerciseByIDRow, error) {
	return s.q.GetExerciseByID(ctx, id)
}

func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}

func (s *QueriesStore) GetAPIKeyByParentID(ctx context.Context, parentID pgtype.UUID) (queries.ApiKey, error) {
	return s.q.GetAPIKeyByParentID(ctx, parentID)
}
