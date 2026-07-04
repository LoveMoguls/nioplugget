package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type sentMessage struct {
	chatID int64
	text   string
	kb     *InlineKeyboardMarkup
}

type fakeSender struct {
	messages []sentMessage
	actions  []string
	answered []string
}

func (f *fakeSender) SendMessage(_ context.Context, chatID int64, text string, kb *InlineKeyboardMarkup) error {
	f.messages = append(f.messages, sentMessage{chatID, text, kb})
	return nil
}
func (f *fakeSender) SendChatAction(_ context.Context, _ int64, action string) error {
	f.actions = append(f.actions, action)
	return nil
}
func (f *fakeSender) AnswerCallbackQuery(_ context.Context, id string) error {
	f.answered = append(f.answered, id)
	return nil
}

// fakeStore: embedded Store gör att oimplementerade metoder panikar i test.
type fakeStore struct {
	Store
	linkCodes  map[string]queries.TelegramLinkCode
	links      map[int64]queries.TelegramLink // key: telegram_user_id
	tgSessions map[int64]queries.TelegramSession
	students   map[string]queries.Student // key: uuidToString(id)
	usedCodes  []string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		linkCodes:  map[string]queries.TelegramLinkCode{},
		links:      map[int64]queries.TelegramLink{},
		tgSessions: map[int64]queries.TelegramSession{},
		students:   map[string]queries.Student{},
	}
}

func (f *fakeStore) GetTelegramLinkCode(_ context.Context, code string) (queries.TelegramLinkCode, error) {
	lc, ok := f.linkCodes[code]
	if !ok {
		return queries.TelegramLinkCode{}, errors.New("not found")
	}
	return lc, nil
}
func (f *fakeStore) MarkTelegramLinkCodeUsed(_ context.Context, code string) error {
	f.usedCodes = append(f.usedCodes, code)
	return nil
}
func (f *fakeStore) UpsertTelegramLink(_ context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error) {
	link := queries.TelegramLink{StudentID: arg.StudentID, TelegramUserID: arg.TelegramUserID, ChatID: arg.ChatID}
	f.links[arg.TelegramUserID] = link
	return link, nil
}
func (f *fakeStore) GetTelegramLinkByTelegramUserID(_ context.Context, id int64) (queries.TelegramLink, error) {
	link, ok := f.links[id]
	if !ok {
		return queries.TelegramLink{}, errors.New("not found")
	}
	return link, nil
}
func (f *fakeStore) GetTelegramSession(_ context.Context, chatID int64) (queries.TelegramSession, error) {
	s, ok := f.tgSessions[chatID]
	if !ok {
		return queries.TelegramSession{}, errors.New("not found")
	}
	return s, nil
}
func (f *fakeStore) UpsertTelegramSession(_ context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error) {
	s := queries.TelegramSession{ChatID: arg.ChatID, StudentID: arg.StudentID, State: arg.State, ActiveSessionID: arg.ActiveSessionID}
	f.tgSessions[arg.ChatID] = s
	return s, nil
}
func (f *fakeStore) GetStudentByID(_ context.Context, id pgtype.UUID) (queries.Student, error) {
	s, ok := f.students[uuidToString(id)]
	if !ok {
		return queries.Student{}, errors.New("not found")
	}
	return s, nil
}
func (f *fakeStore) ListDueReviews(_ context.Context, _ pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return nil, nil
}

func testUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}

func TestStartWithValidCodeLinksAccount(t *testing.T) {
	store := newFakeStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.students[uuidToString(studentID)] = queries.Student{ID: studentID, Name: "Nio"}
	store.linkCodes["ABCD2345"] = queries.TelegramLinkCode{
		Code:      "ABCD2345",
		StudentID: studentID,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 777}, Chat: Chat{ID: 555}, Text: "/start ABCD2345",
	}})

	if _, ok := store.links[777]; !ok {
		t.Fatal("expected link to be created")
	}
	if len(store.usedCodes) != 1 || store.usedCodes[0] != "ABCD2345" {
		t.Errorf("code not marked used: %v", store.usedCodes)
	}
	if len(sender.messages) < 2 {
		t.Fatalf("expected welcome + menu, got %d messages", len(sender.messages))
	}
	if !strings.Contains(sender.messages[0].text, "Nio") {
		t.Errorf("welcome should greet by name: %q", sender.messages[0].text)
	}
	menu := sender.messages[len(sender.messages)-1]
	if menu.kb == nil || len(menu.kb.InlineKeyboard) != 4 {
		t.Errorf("expected 4-row main menu keyboard, got %+v", menu.kb)
	}
}

func TestStartWithExpiredCodeRejects(t *testing.T) {
	store := newFakeStore()
	store.linkCodes["OLDCODE2"] = queries.TelegramLinkCode{
		Code:      "OLDCODE2",
		StudentID: testUUID("11111111-1111-1111-1111-111111111111"),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
	}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 777}, Chat: Chat{ID: 555}, Text: "/start OLDCODE2",
	}})

	if len(store.links) != 0 {
		t.Error("expired code must not link")
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "ogiltig") {
		t.Errorf("expected rejection, got %v", sender.messages)
	}
}

func TestUnlinkedUserIsRejected(t *testing.T) {
	sender := &fakeSender{}
	bot := NewBot(sender, newFakeStore(), nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 999}, Chat: Chat{ID: 888}, Text: "hej boten",
	}})

	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "kopplingslänk") {
		t.Errorf("expected unlinked reply, got %v", sender.messages)
	}
}

func TestCallbackIsAnswered(t *testing.T) {
	store := newFakeStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{CallbackQuery: &CallbackQuery{
		ID: "cb1", From: User{ID: 777}, Message: &Message{Chat: Chat{ID: 555}}, Data: "menu",
	}})

	if len(sender.answered) != 1 || sender.answered[0] != "cb1" {
		t.Errorf("callback not answered: %v", sender.answered)
	}
	if len(sender.messages) != 1 || sender.messages[0].kb == nil {
		t.Errorf("expected main menu, got %v", sender.messages)
	}
}
