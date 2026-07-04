package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// dialogStore extends fakeStore with session/message/exercise state.
type dialogStore struct {
	*fakeStore
	sessions  map[string]queries.GetSessionByIDRow
	messages  []queries.Message
	exercises map[string]queries.GetExerciseByIDRow
	apiKeys   map[string]queries.ApiKey // key: uuidToString(parentID)
	ended     *queries.EndSessionParams
	schedule  *queries.UpsertReviewScheduleParams
}

func newDialogStore() *dialogStore {
	return &dialogStore{
		fakeStore: newFakeStore(),
		sessions:  map[string]queries.GetSessionByIDRow{},
		exercises: map[string]queries.GetExerciseByIDRow{},
		apiKeys:   map[string]queries.ApiKey{},
	}
}

func (d *dialogStore) CreateSession(_ context.Context, arg queries.CreateSessionParams) (queries.CreateSessionRow, error) {
	id := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(id)] = queries.GetSessionByIDRow{ID: id, StudentID: arg.StudentID, ExerciseID: arg.ExerciseID}
	return queries.CreateSessionRow{ID: id, StudentID: arg.StudentID, ExerciseID: arg.ExerciseID}, nil
}
func (d *dialogStore) GetSessionByID(_ context.Context, id pgtype.UUID) (queries.GetSessionByIDRow, error) {
	s, ok := d.sessions[uuidToString(id)]
	if !ok {
		return queries.GetSessionByIDRow{}, errors.New("not found")
	}
	return s, nil
}
func (d *dialogStore) CreateMessage(_ context.Context, arg queries.CreateMessageParams) (queries.Message, error) {
	m := queries.Message{SessionID: arg.SessionID, Role: arg.Role, Content: arg.Content}
	d.messages = append(d.messages, m)
	return m, nil
}
func (d *dialogStore) ListMessagesBySessionID(_ context.Context, _ pgtype.UUID) ([]queries.Message, error) {
	return d.messages, nil
}
func (d *dialogStore) GetExerciseByID(_ context.Context, id pgtype.UUID) (queries.GetExerciseByIDRow, error) {
	e, ok := d.exercises[uuidToString(id)]
	if !ok {
		return queries.GetExerciseByIDRow{}, errors.New("not found")
	}
	return e, nil
}
func (d *dialogStore) GetAPIKeyByParentID(_ context.Context, parentID pgtype.UUID) (queries.ApiKey, error) {
	k, ok := d.apiKeys[uuidToString(parentID)]
	if !ok {
		return queries.ApiKey{}, errors.New("not found")
	}
	return k, nil
}
func (d *dialogStore) EndSession(_ context.Context, arg queries.EndSessionParams) (queries.EndSessionRow, error) {
	d.ended = &arg
	s := d.sessions[uuidToString(arg.ID)]
	s.EndedAt = pgtype.Timestamptz{Valid: true}
	d.sessions[uuidToString(arg.ID)] = s
	return queries.EndSessionRow{ID: arg.ID}, nil
}
func (d *dialogStore) GetReviewSchedule(_ context.Context, _ queries.GetReviewScheduleParams) (queries.ReviewSchedule, error) {
	return queries.ReviewSchedule{}, errors.New("not found")
}
func (d *dialogStore) UpsertReviewSchedule(_ context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error) {
	d.schedule = &arg
	return queries.ReviewSchedule{}, nil
}

func setupLinkedStudent(d *dialogStore) queries.TelegramLink {
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	parentID := testUUID("44444444-4444-4444-4444-444444444444")
	d.students[uuidToString(studentID)] = queries.Student{ID: studentID, ParentID: parentID, Name: "Nio"}
	link := queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	d.links[777] = link
	return link
}

func TestStartExerciseCreatesSessionAndState(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes", Description: "Förklara fotosyntesen", SystemPrompt: "Du är en sokratisk lärare."}
	sender := &fakeSender{}
	bot := NewBot(sender, d, nil)

	bot.startExercise(context.Background(), link, uuidToString(exID))

	tg := d.tgSessions[555]
	if tg.State != "in_session" || !tg.ActiveSessionID.Valid {
		t.Fatalf("telegram session not in_session: %+v", tg)
	}
	if len(sender.messages) == 0 || !strings.Contains(sender.messages[0].text, "Fotosyntes") {
		t.Errorf("intro should mention exercise title: %v", sender.messages)
	}
}

func TestDialogMessageCallsClaudeWithTelegramSuffix(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes", SystemPrompt: "Sokratisk lärare."}
	sessionID := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(sessionID)] = queries.GetSessionByIDRow{ID: sessionID, StudentID: link.StudentID, ExerciseID: exID}
	// Fake encryptable API key path: bypass encryption by stubbing decryption
	// via a bot whose encSvc is replaced — instead we stub complete/score and
	// return the key straight from a fake encSvc-less path is not possible,
	// so store an encrypted key with a real EncryptionService:
	sender := &fakeSender{}
	bot := NewBot(sender, d, testEncSvc(t, d, link))

	var gotSystem string
	bot.complete = func(_ context.Context, _, systemPrompt string, _ []anthropic.MessageParam) (string, error) {
		gotSystem = systemPrompt
		return "Bra fråga! Vad tror du själv?", nil
	}

	tgSession := queries.TelegramSession{ChatID: 555, StudentID: link.StudentID, State: "in_session", ActiveSessionID: sessionID}
	bot.handleDialogMessage(context.Background(), link, tgSession, "Vad är klorofyll?")

	if !strings.Contains(gotSystem, "LaTeX") {
		t.Errorf("system prompt must include telegram suffix, got %q", gotSystem)
	}
	if len(d.messages) != 2 || d.messages[0].Role != "user" || d.messages[1].Role != "assistant" {
		t.Errorf("expected saved user+assistant messages, got %+v", d.messages)
	}
	if len(sender.actions) == 0 || sender.actions[0] != "typing" {
		t.Errorf("expected typing action, got %v", sender.actions)
	}
	last := sender.messages[len(sender.messages)-1]
	if !strings.Contains(last.text, "Vad tror du själv?") || last.kb == nil {
		t.Errorf("reply should carry end keyboard: %+v", last)
	}
}

func TestEndActiveSessionScoresAndSchedules(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes"}
	sessionID := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(sessionID)] = queries.GetSessionByIDRow{ID: sessionID, StudentID: link.StudentID, ExerciseID: exID}
	d.tgSessions[555] = queries.TelegramSession{ChatID: 555, StudentID: link.StudentID, State: "in_session", ActiveSessionID: sessionID}
	sender := &fakeSender{}
	bot := NewBot(sender, d, testEncSvc(t, d, link))
	bot.score = func(_ context.Context, _, _ string, _ []queries.Message) (*chat.ScoreResult, error) {
		return &chat.ScoreResult{Score: 5, Summary: "Toppen!", Feedback: "Snyggt!"}, nil
	}

	bot.endActiveSession(context.Background(), link)

	if d.ended == nil || d.ended.Score.Int32 != 5 {
		t.Fatalf("session not ended with score: %+v", d.ended)
	}
	if d.schedule == nil {
		t.Fatal("SM-2 schedule not updated")
	}
	if d.tgSessions[555].State != "menu" {
		t.Errorf("telegram session should reset to menu, got %q", d.tgSessions[555].State)
	}
	var starsMsg string
	for _, m := range sender.messages {
		if strings.Contains(m.text, "⭐") {
			starsMsg = m.text
		}
	}
	if !strings.Contains(starsMsg, "⭐⭐⭐") || !strings.Contains(starsMsg, "+30 XP") {
		t.Errorf("expected 3 stars +30 XP, got %q", starsMsg)
	}
}

// testEncSvc creates a real EncryptionService with a test key and stores an
// encrypted API key for the linked student's parent.
func testEncSvc(t *testing.T, d *dialogStore, link queries.TelegramLink) *apikey.EncryptionService {
	t.Helper()
	// apikey.NewEncryptionService expects a hex-encoded 32-byte key (64 hex
	// chars); verified in internal/apikey/encrypt.go.
	svc, err := apikey.NewEncryptionService(strings.Repeat("a", 64))
	if err != nil {
		t.Fatalf("enc svc: %v", err)
	}
	student := d.students[uuidToString(link.StudentID)]
	enc, err := svc.Encrypt([]byte("sk-ant-test"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	d.apiKeys[uuidToString(student.ParentID)] = queries.ApiKey{ParentID: student.ParentID, EncryptedKey: enc}
	return svc
}
