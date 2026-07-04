package telegram

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Store is the database access the Telegram bot needs.
type Store interface {
	chat.ChatStore

	CreateTelegramLinkCode(ctx context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error)
	GetTelegramLinkCode(ctx context.Context, code string) (queries.TelegramLinkCode, error)
	MarkTelegramLinkCodeUsed(ctx context.Context, code string) error
	UpsertTelegramLink(ctx context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error)
	GetTelegramLinkByTelegramUserID(ctx context.Context, telegramUserID int64) (queries.TelegramLink, error)
	ListTelegramLinks(ctx context.Context) ([]queries.TelegramLink, error)
	GetTelegramSession(ctx context.Context, chatID int64) (queries.TelegramSession, error)
	UpsertTelegramSession(ctx context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error)
	TryInsertTelegramReminder(ctx context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error)

	ListSubjects(ctx context.Context) ([]queries.ListSubjectsRow, error)
	GetSubjectBySlug(ctx context.Context, slug string) (queries.GetSubjectBySlugRow, error)
	ListTopicsBySubjectID(ctx context.Context, subjectID pgtype.UUID) ([]queries.ListTopicsBySubjectIDRow, error)
	GetTopicBySlug(ctx context.Context, arg queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error)
	ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error)
	ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListPublishedChallengesByParentIDRow, error)
	ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error)
	ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error)
	GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error)
}

// QueriesStore implements Store using sqlc-generated queries.
// chat.QueriesStore provides the chat.ChatStore subset.
type QueriesStore struct {
	*chat.QueriesStore
	q *queries.Queries
}

func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{QueriesStore: chat.NewQueriesStore(q), q: q}
}

func (s *QueriesStore) CreateTelegramLinkCode(ctx context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error) {
	return s.q.CreateTelegramLinkCode(ctx, arg)
}

func (s *QueriesStore) GetTelegramLinkCode(ctx context.Context, code string) (queries.TelegramLinkCode, error) {
	return s.q.GetTelegramLinkCode(ctx, code)
}

func (s *QueriesStore) MarkTelegramLinkCodeUsed(ctx context.Context, code string) error {
	return s.q.MarkTelegramLinkCodeUsed(ctx, code)
}

func (s *QueriesStore) UpsertTelegramLink(ctx context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error) {
	return s.q.UpsertTelegramLink(ctx, arg)
}

func (s *QueriesStore) GetTelegramLinkByTelegramUserID(ctx context.Context, telegramUserID int64) (queries.TelegramLink, error) {
	return s.q.GetTelegramLinkByTelegramUserID(ctx, telegramUserID)
}

func (s *QueriesStore) ListTelegramLinks(ctx context.Context) ([]queries.TelegramLink, error) {
	return s.q.ListTelegramLinks(ctx)
}

func (s *QueriesStore) GetTelegramSession(ctx context.Context, chatID int64) (queries.TelegramSession, error) {
	return s.q.GetTelegramSession(ctx, chatID)
}

func (s *QueriesStore) UpsertTelegramSession(ctx context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error) {
	return s.q.UpsertTelegramSession(ctx, arg)
}

func (s *QueriesStore) TryInsertTelegramReminder(ctx context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error) {
	return s.q.TryInsertTelegramReminder(ctx, arg)
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

func (s *QueriesStore) GetTopicBySlug(ctx context.Context, arg queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error) {
	return s.q.GetTopicBySlug(ctx, arg)
}

func (s *QueriesStore) ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error) {
	return s.q.ListExercisesByTopicID(ctx, topicID)
}

func (s *QueriesStore) ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListPublishedChallengesByParentIDRow, error) {
	return s.q.ListPublishedChallengesByParentID(ctx, parentID)
}

func (s *QueriesStore) ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error) {
	return s.q.ListChallengeExercisesWithProgress(ctx, arg)
}

func (s *QueriesStore) ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return s.q.ListDueReviews(ctx, studentID)
}

func (s *QueriesStore) GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error) {
	return s.q.GetStudentProgressBySubject(ctx, studentID)
}
