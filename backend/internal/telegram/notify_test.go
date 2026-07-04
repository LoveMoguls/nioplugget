package telegram

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type notifyStore struct {
	*fakeStore
	due       map[string][]queries.ListDueReviewsRow // key: uuidToString(studentID)
	reminders map[string]bool                        // key: studentID+date
	challenge *queries.Challenge
}

func newNotifyStore() *notifyStore {
	return &notifyStore{fakeStore: newFakeStore(), due: map[string][]queries.ListDueReviewsRow{}, reminders: map[string]bool{}}
}

func (n *notifyStore) ListTelegramLinks(_ context.Context) ([]queries.TelegramLink, error) {
	var links []queries.TelegramLink
	for _, l := range n.links {
		links = append(links, l)
	}
	return links, nil
}
func (n *notifyStore) ListDueReviews(_ context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return n.due[uuidToString(studentID)], nil
}
func (n *notifyStore) TryInsertTelegramReminder(_ context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error) {
	key := uuidToString(arg.StudentID) + arg.SentOn.Time.Format("2006-01-02")
	if n.reminders[key] {
		return 0, nil
	}
	n.reminders[key] = true
	return 1, nil
}

func (n *notifyStore) GetChallengeByID(_ context.Context, _ pgtype.UUID) (queries.Challenge, error) {
	if n.challenge == nil {
		return queries.Challenge{}, errors.New("not found")
	}
	return *n.challenge, nil
}

func TestReminderSentOncePerDayInsideWindow(t *testing.T) {
	store := newNotifyStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	store.due[uuidToString(studentID)] = []queries.ListDueReviewsRow{{ExerciseTitle: "Fotosyntes"}}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	at16 := time.Date(2026, 7, 4, 16, 0, 0, 0, time.Local)
	bot.sendDueReminders(context.Background(), at16)
	bot.sendDueReminders(context.Background(), at16.Add(time.Hour))

	if len(sender.messages) != 1 {
		t.Fatalf("expected exactly 1 reminder, got %d", len(sender.messages))
	}
	if sender.messages[0].kb == nil {
		t.Error("reminder should carry a 'Kör nu' button")
	}
}

func TestNoReminderOutsideWindow(t *testing.T) {
	store := newNotifyStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	store.due[uuidToString(studentID)] = []queries.ListDueReviewsRow{{ExerciseTitle: "Fotosyntes"}}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.sendDueReminders(context.Background(), time.Date(2026, 7, 4, 9, 0, 0, 0, time.Local))

	if len(sender.messages) != 0 {
		t.Errorf("expected no reminder at 09:00, got %v", sender.messages)
	}
}

func TestChallengePublishedNotifiesFamilyChildren(t *testing.T) {
	store := newNotifyStore()
	parentID := testUUID("44444444-4444-4444-4444-444444444444")
	otherParent := testUUID("55555555-5555-5555-5555-555555555555")
	kid1 := testUUID("11111111-1111-1111-1111-111111111111")
	kid2 := testUUID("22222222-2222-2222-2222-222222222222")
	store.students[uuidToString(kid1)] = queries.Student{ID: kid1, ParentID: parentID}
	store.students[uuidToString(kid2)] = queries.Student{ID: kid2, ParentID: otherParent}
	store.links[701] = queries.TelegramLink{StudentID: kid1, TelegramUserID: 701, ChatID: 501}
	store.links[702] = queries.TelegramLink{StudentID: kid2, TelegramUserID: 702, ChatID: 502}
	chalID := testUUID("66666666-6666-6666-6666-666666666666")
	store.challenge = &queries.Challenge{ID: chalID, ParentID: parentID, Title: "Glosor v.28", CoverEmoji: "📖", Published: true}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.ChallengePublished(context.Background(), chalID)

	if len(sender.messages) != 1 || sender.messages[0].chatID != 501 {
		t.Fatalf("expected exactly one notification to chat 501, got %v", sender.messages)
	}
}
