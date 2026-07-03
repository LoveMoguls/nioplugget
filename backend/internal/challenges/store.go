// backend/internal/challenges/store.go
package challenges

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ChallengeStore abstracts DB access for challenges.
type ChallengeStore interface {
	CreateChallenge(ctx context.Context, arg queries.CreateChallengeParams) (queries.Challenge, error)
	CreateChallengeExercise(ctx context.Context, arg queries.CreateChallengeExerciseParams) (queries.ChallengeExercise, error)
	ListChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Challenge, error)
	ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Challenge, error)
	PublishChallenge(ctx context.Context, arg queries.PublishChallengeParams) (queries.Challenge, error)
	GetChallengeByID(ctx context.Context, id pgtype.UUID) (queries.Challenge, error)
	ListChallengeExercisesByChallengeID(ctx context.Context, challengeID pgtype.UUID) ([]queries.ChallengeExercise, error)
	ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error)
	DeleteChallenge(ctx context.Context, arg queries.DeleteChallengeParams) error
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
	GetAPIKeyByParentID(ctx context.Context, parentID pgtype.UUID) (queries.ApiKey, error)
}

// QueriesStore implements ChallengeStore using sqlc-generated queries.
type QueriesStore struct {
	q *queries.Queries
}

func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{q: q}
}

func (s *QueriesStore) CreateChallenge(ctx context.Context, arg queries.CreateChallengeParams) (queries.Challenge, error) {
	return s.q.CreateChallenge(ctx, arg)
}

func (s *QueriesStore) CreateChallengeExercise(ctx context.Context, arg queries.CreateChallengeExerciseParams) (queries.ChallengeExercise, error) {
	return s.q.CreateChallengeExercise(ctx, arg)
}

func (s *QueriesStore) ListChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Challenge, error) {
	return s.q.ListChallengesByParentID(ctx, parentID)
}

func (s *QueriesStore) ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Challenge, error) {
	return s.q.ListPublishedChallengesByParentID(ctx, parentID)
}

func (s *QueriesStore) PublishChallenge(ctx context.Context, arg queries.PublishChallengeParams) (queries.Challenge, error) {
	return s.q.PublishChallenge(ctx, arg)
}

func (s *QueriesStore) GetChallengeByID(ctx context.Context, id pgtype.UUID) (queries.Challenge, error) {
	return s.q.GetChallengeByID(ctx, id)
}

func (s *QueriesStore) ListChallengeExercisesByChallengeID(ctx context.Context, challengeID pgtype.UUID) ([]queries.ChallengeExercise, error) {
	return s.q.ListChallengeExercisesByChallengeID(ctx, challengeID)
}

func (s *QueriesStore) ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error) {
	return s.q.ListChallengeExercisesWithProgress(ctx, arg)
}

func (s *QueriesStore) DeleteChallenge(ctx context.Context, arg queries.DeleteChallengeParams) error {
	return s.q.DeleteChallenge(ctx, arg)
}

func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}

func (s *QueriesStore) GetAPIKeyByParentID(ctx context.Context, parentID pgtype.UUID) (queries.ApiKey, error) {
	return s.q.GetAPIKeyByParentID(ctx, parentID)
}

// resolveParentID returns parent_id and created_by_role for both parent and child callers.
func resolveParentID(ctx context.Context, store ChallengeStore, userID, role string) (pgtype.UUID, string, error) {
	var parentID pgtype.UUID
	if role == "parent" {
		if err := parentID.Scan(userID); err != nil {
			return pgtype.UUID{}, "", fmt.Errorf("invalid parent ID")
		}
		return parentID, "parent", nil
	}
	var studentID pgtype.UUID
	if err := studentID.Scan(userID); err != nil {
		return pgtype.UUID{}, "", fmt.Errorf("invalid student ID")
	}
	student, err := store.GetStudentByID(ctx, studentID)
	if err != nil {
		return pgtype.UUID{}, "", fmt.Errorf("student not found")
	}
	return student.ParentID, "child", nil
}
