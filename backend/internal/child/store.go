package child

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// DBTX is the interface for direct database access (matches pgxpool.Pool and pgx.Conn).
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) interface {
		Scan(dest ...any) error
	}
}

// QueriesStore wraps *queries.Queries and a DBTX to implement ChildQuerier.
// It delegates all standard queries and adds UpdateStudentInvite which
// is not in the sqlc-generated code.
type QueriesStore struct {
	q  *queries.Queries
	db queries.DBTX
}

// NewQueriesStore creates a QueriesStore backed by sqlc-generated queries.
// The db parameter is used for the UpdateStudentInvite raw query.
func NewQueriesStore(q *queries.Queries, db queries.DBTX) *QueriesStore {
	return &QueriesStore{q: q, db: db}
}

func (s *QueriesStore) CreateStudent(ctx context.Context, arg queries.CreateStudentParams) (queries.Student, error) {
	return s.q.CreateStudent(ctx, arg)
}

func (s *QueriesStore) GetStudentsByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Student, error) {
	return s.q.GetStudentsByParentID(ctx, parentID)
}

func (s *QueriesStore) ActivateStudent(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error) {
	return s.q.ActivateStudent(ctx, arg)
}

func (s *QueriesStore) GetStudentByNameAndParent(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error) {
	return s.q.GetStudentByNameAndParent(ctx, arg)
}

func (s *QueriesStore) GetParentByEmail(ctx context.Context, email string) (queries.Parent, error) {
	return s.q.GetParentByEmail(ctx, email)
}

func (s *QueriesStore) ListStudentNamesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListStudentNamesByParentIDRow, error) {
	return s.q.ListStudentNamesByParentID(ctx, parentID)
}

func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}

// updateStudentInviteSQL updates an existing student's invite code and expiry.
// This query is not in the sqlc-generated code so we execute it directly.
const updateStudentInviteSQL = `
UPDATE students
SET invite_code = $2,
    invite_expires_at = $3
WHERE id = $1
RETURNING id, parent_id, name, pin_hash, invite_code, invite_expires_at, activated_at, created_at
`

// UpdateStudentInvite refreshes the invite code and expiry for an existing student.
func (s *QueriesStore) UpdateStudentInvite(ctx context.Context, arg UpdateStudentInviteParams) (queries.Student, error) {
	row := s.db.QueryRow(ctx, updateStudentInviteSQL, arg.ID, arg.InviteCode, arg.InviteExpiresAt)
	var st queries.Student
	err := row.Scan(
		&st.ID,
		&st.ParentID,
		&st.Name,
		&st.PinHash,
		&st.InviteCode,
		&st.InviteExpiresAt,
		&st.ActivatedAt,
		&st.CreatedAt,
	)
	return st, err
}
