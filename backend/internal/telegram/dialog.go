package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/challenges"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// telegramPromptSuffix is appended to every system prompt in the Telegram
// channel: Telegram cannot render LaTeX.
const telegramPromptSuffix = "\n\nVIKTIGT FÖR DENNA KANAL: Du chattar via Telegram. Skriv aldrig LaTeX. Skriv matematik som enkel text med unicode, t.ex. x² + 3x = 10, √16, ½."

func endKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{{Text: "✅ Klar — avsluta och rätta", CallbackData: "end"}},
	}}
}

func (b *Bot) startExercise(ctx context.Context, link queries.TelegramLink, exerciseID string) {
	exercise, err := b.store.GetExerciseByID(ctx, parseUUID(exerciseID))
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	session, err := b.store.CreateSession(ctx, queries.CreateSessionParams{
		StudentID:  link.StudentID,
		ExerciseID: exercise.ID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create session from telegram")
		b.send(ctx, link.ChatID, "Kunde inte starta övningen. Försök igen.", nil)
		return
	}
	b.setActiveSession(ctx, link, session.ID)
	b.send(ctx, link.ChatID, fmt.Sprintf("📚 %s\n\n%s\n\nSkriv ditt svar eller din fundering — jag guidar dig. Skriv /avsluta när du känner dig klar.", exercise.Title, exercise.Description), endKeyboard())
}

func (b *Bot) startChallengeExercise(ctx context.Context, link queries.TelegramLink, chexIDStr string) {
	chexID := parseUUID(chexIDStr)
	chex, err := b.store.GetChallengeExerciseByID(ctx, chexID)
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	challenge, err := b.store.GetChallengeByID(ctx, chex.ChallengeID)
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	student, err := b.store.GetStudentByID(ctx, link.StudentID)
	if err != nil || challenge.ParentID != student.ParentID || !challenge.Published {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	session, err := b.store.CreateChallengeSession(ctx, queries.CreateChallengeSessionParams{
		StudentID:           link.StudentID,
		ChallengeExerciseID: chexID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create challenge session from telegram")
		b.send(ctx, link.ChatID, "Kunde inte starta övningen. Försök igen.", nil)
		return
	}
	b.setActiveSession(ctx, link, session.ID)
	b.send(ctx, link.ChatID, fmt.Sprintf("🏆 %s\n\n%s\n\nSkriv ditt svar eller din fundering — jag guidar dig. Skriv /avsluta när du känner dig klar.", chex.Title, chex.Description), endKeyboard())
}

func (b *Bot) setActiveSession(ctx context.Context, link queries.TelegramLink, sessionID pgtype.UUID) {
	if _, err := b.store.UpsertTelegramSession(ctx, queries.UpsertTelegramSessionParams{
		ChatID:          link.ChatID,
		StudentID:       link.StudentID,
		State:           "in_session",
		ActiveSessionID: sessionID,
	}); err != nil {
		log.Error().Err(err).Msg("failed to set telegram session state")
	}
}

func (b *Bot) clearTelegramSession(ctx context.Context, link queries.TelegramLink) {
	if _, err := b.store.UpsertTelegramSession(ctx, queries.UpsertTelegramSessionParams{
		ChatID:    link.ChatID,
		StudentID: link.StudentID,
		State:     "menu",
	}); err != nil {
		log.Error().Err(err).Msg("failed to reset telegram session")
	}
}

func (b *Bot) decryptedAPIKey(ctx context.Context, studentID pgtype.UUID) (string, error) {
	student, err := b.store.GetStudentByID(ctx, studentID)
	if err != nil {
		return "", fmt.Errorf("student not found: %w", err)
	}
	rec, err := b.store.GetAPIKeyByParentID(ctx, student.ParentID)
	if err != nil {
		return "", fmt.Errorf("API key not found: %w", err)
	}
	plaintext, err := b.encSvc.Decrypt(rec.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(plaintext), nil
}

func (b *Bot) handleDialogMessage(ctx context.Context, link queries.TelegramLink, tgSession queries.TelegramSession, text string) {
	sessionID := tgSession.ActiveSessionID
	session, err := b.store.GetSessionByID(ctx, sessionID)
	if err != nil || session.EndedAt.Valid {
		b.clearTelegramSession(ctx, link)
		b.sendMainMenu(ctx, link)
		return
	}
	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "user", Content: text,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save user message")
		b.send(ctx, link.ChatID, "Något gick fel. Försök igen.", nil)
		return
	}
	var systemPrompt string
	if session.ExerciseID.Valid {
		ex, err := b.store.GetExerciseByID(ctx, session.ExerciseID)
		if err != nil {
			b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
			return
		}
		systemPrompt = ex.SystemPrompt
	} else {
		chex, err := b.store.GetChallengeExerciseByID(ctx, session.ChallengeExerciseID)
		if err != nil {
			b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
			return
		}
		systemPrompt = chex.SystemPrompt
	}
	systemPrompt += telegramPromptSuffix

	apiKey, err := b.decryptedAPIKey(ctx, link.StudentID)
	if err != nil {
		log.Error().Err(err).Msg("api key unavailable for telegram chat")
		b.send(ctx, link.ChatID, "Förälderns API-nyckel saknas eller är ogiltig.", nil)
		return
	}
	msgs, err := b.store.ListMessagesBySessionID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages")
		b.send(ctx, link.ChatID, "Något gick fel. Försök igen.", nil)
		return
	}
	history := chat.BuildMessageHistory(msgs, chat.DefaultWindowSize)
	if err := b.api.SendChatAction(ctx, link.ChatID, "typing"); err != nil {
		log.Error().Err(err).Msg("failed to send typing action")
	}
	reply, err := b.complete(ctx, apiKey, systemPrompt, history)
	if err != nil {
		log.Error().Err(err).Msg("claude completion failed")
		b.send(ctx, link.ChatID, "AI-tjänsten svarade inte. Försök igen om en stund.", nil)
		return
	}
	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "assistant", Content: reply,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save assistant message")
	}
	b.send(ctx, link.ChatID, reply, endKeyboard())
}

func (b *Bot) endActiveSession(ctx context.Context, link queries.TelegramLink) {
	tgSession, err := b.store.GetTelegramSession(ctx, link.ChatID)
	if err != nil || tgSession.State != "in_session" || !tgSession.ActiveSessionID.Valid {
		b.sendMainMenu(ctx, link)
		return
	}
	sessionID := tgSession.ActiveSessionID
	session, err := b.store.GetSessionByID(ctx, sessionID)
	if err != nil || session.EndedAt.Valid {
		b.clearTelegramSession(ctx, link)
		b.sendMainMenu(ctx, link)
		return
	}
	msgs, err := b.store.ListMessagesBySessionID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages for scoring")
	}
	var title string
	if session.ExerciseID.Valid {
		if ex, err := b.store.GetExerciseByID(ctx, session.ExerciseID); err == nil {
			title = ex.Title
		}
	} else if chex, err := b.store.GetChallengeExerciseByID(ctx, session.ChallengeExerciseID); err == nil {
		title = chex.Title
	}

	result := &chat.ScoreResult{Score: 3, Summary: "Passet avslutades utan bedömning.", Feedback: "Bra jobbat att du övade!"}
	if apiKey, err := b.decryptedAPIKey(ctx, link.StudentID); err == nil {
		if scored, err := b.score(ctx, apiKey, title, msgs); err == nil && scored != nil {
			result = scored
		}
	}

	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "assistant", Content: result.Summary,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save summary message")
	}
	if _, err := b.store.EndSession(ctx, queries.EndSessionParams{
		ID:      sessionID,
		Score:   pgtype.Int4{Int32: int32(result.Score), Valid: true},
		Summary: pgtype.Text{String: result.Summary, Valid: true},
	}); err != nil {
		log.Error().Err(err).Msg("failed to end session")
		b.send(ctx, link.ChatID, "Kunde inte avsluta passet. Försök igen.", nil)
		return
	}
	if session.ExerciseID.Valid {
		if err := chat.UpdateReviewSchedule(ctx, b.store, link.StudentID, session.ExerciseID, result.Score, time.Now()); err != nil {
			log.Error().Err(err).Msg("failed to update review schedule")
		}
	}
	b.clearTelegramSession(ctx, link)

	stars := challenges.ScoreToStars(result.Score)
	xp := stars * 10
	b.send(ctx, link.ChatID, fmt.Sprintf("%s %d stjärnor! +%d XP\n\n%s\n\n%s",
		strings.Repeat("⭐", stars), stars, xp, result.Summary, result.Feedback), nil)
	b.sendMainMenu(ctx, link)
}
