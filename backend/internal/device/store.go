package device

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Store is the database access the device/profile endpoints need.
type Store interface {
	GetFamilySettings(ctx context.Context) (queries.FamilySetting, error)
	UpsertFamilyCode(ctx context.Context, codeHash string) (queries.FamilySetting, error)
	ListParents(ctx context.Context) ([]queries.ListParentsRow, error)
	ListAllStudents(ctx context.Context) ([]queries.ListAllStudentsRow, error)
	GetParentByID(ctx context.Context, id pgtype.UUID) (queries.Parent, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
}

// QueriesStore implements Store with sqlc-generated queries.
type QueriesStore struct{ q *queries.Queries }

func NewQueriesStore(q *queries.Queries) *QueriesStore { return &QueriesStore{q: q} }

func (s *QueriesStore) GetFamilySettings(ctx context.Context) (queries.FamilySetting, error) {
	return s.q.GetFamilySettings(ctx)
}
func (s *QueriesStore) UpsertFamilyCode(ctx context.Context, codeHash string) (queries.FamilySetting, error) {
	return s.q.UpsertFamilyCode(ctx, codeHash)
}
func (s *QueriesStore) ListParents(ctx context.Context) ([]queries.ListParentsRow, error) {
	return s.q.ListParents(ctx)
}
func (s *QueriesStore) ListAllStudents(ctx context.Context) ([]queries.ListAllStudentsRow, error) {
	return s.q.ListAllStudents(ctx)
}
func (s *QueriesStore) GetParentByID(ctx context.Context, id pgtype.UUID) (queries.Parent, error) {
	return s.q.GetParentByID(ctx, id)
}
func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}
