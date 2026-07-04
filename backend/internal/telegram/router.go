package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

const unlinkedReply = "Hej! Jag är Nioplugget-boten. Hämta en kopplingslänk på nioplugget-webben (knappen \"Koppla Telegram\") för att komma igång."

// HandleUpdate routes a single Telegram update. Errors are logged, never fatal.
func (b *Bot) HandleUpdate(ctx context.Context, upd Update) {
	switch {
	case upd.CallbackQuery != nil:
		b.handleCallback(ctx, *upd.CallbackQuery)
	case upd.Message != nil && upd.Message.From != nil && upd.Message.Text != "":
		b.handleMessage(ctx, *upd.Message)
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg Message) {
	if strings.HasPrefix(msg.Text, "/start") {
		b.handleStart(ctx, msg)
		return
	}
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, msg.From.ID)
	if err != nil {
		b.send(ctx, msg.Chat.ID, unlinkedReply, nil)
		return
	}
	if msg.Text == "/avsluta" {
		b.endActiveSession(ctx, link)
		return
	}
	tgSession, err := b.store.GetTelegramSession(ctx, link.ChatID)
	if err == nil && tgSession.State == "in_session" && tgSession.ActiveSessionID.Valid {
		b.handleDialogMessage(ctx, link, tgSession, msg.Text)
		return
	}
	b.sendMainMenu(ctx, link)
}

func (b *Bot) handleStart(ctx context.Context, msg Message) {
	chatID := msg.Chat.ID
	parts := strings.Fields(msg.Text)
	if len(parts) == 2 {
		code := strings.ToUpper(parts[1])
		lc, err := b.store.GetTelegramLinkCode(ctx, code)
		if err != nil || lc.UsedAt.Valid || time.Now().After(lc.ExpiresAt.Time) {
			b.send(ctx, chatID, "Koden är ogiltig eller har gått ut. Hämta en ny länk från webben.", nil)
			return
		}
		link, err := b.store.UpsertTelegramLink(ctx, queries.UpsertTelegramLinkParams{
			StudentID:      lc.StudentID,
			TelegramUserID: msg.From.ID,
			ChatID:         chatID,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to link telegram account")
			b.send(ctx, chatID, "Något gick fel vid kopplingen. Försök igen.", nil)
			return
		}
		if err := b.store.MarkTelegramLinkCodeUsed(ctx, code); err != nil {
			log.Error().Err(err).Msg("failed to mark link code used")
		}
		name := ""
		if student, err := b.store.GetStudentByID(ctx, link.StudentID); err == nil {
			name = " " + student.Name
		}
		b.send(ctx, chatID, fmt.Sprintf("Kopplat! 🎉 Hej%s, nu kan du plugga direkt här i Telegram.", name), nil)
		b.sendMainMenu(ctx, link)
		return
	}
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, msg.From.ID)
	if err != nil {
		b.send(ctx, chatID, unlinkedReply, nil)
		return
	}
	b.sendMainMenu(ctx, link)
}

func (b *Bot) handleCallback(ctx context.Context, cb CallbackQuery) {
	if err := b.api.AnswerCallbackQuery(ctx, cb.ID); err != nil {
		log.Error().Err(err).Msg("failed to answer callback query")
	}
	if cb.Message == nil {
		return
	}
	chatID := cb.Message.Chat.ID
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, cb.From.ID)
	if err != nil {
		b.send(ctx, chatID, unlinkedReply, nil)
		return
	}
	data := cb.Data
	switch {
	case data == "menu":
		b.sendMainMenu(ctx, link)
	case data == "study":
		b.sendSubjects(ctx, link)
	case strings.HasPrefix(data, "subj:"):
		b.sendTopics(ctx, link, strings.TrimPrefix(data, "subj:"))
	case strings.HasPrefix(data, "top:"):
		if segs := strings.SplitN(strings.TrimPrefix(data, "top:"), ":", 2); len(segs) == 2 {
			b.sendExercises(ctx, link, segs[0], segs[1])
		}
	case strings.HasPrefix(data, "ex:"):
		b.startExercise(ctx, link, strings.TrimPrefix(data, "ex:"))
	case data == "chals":
		b.sendChallenges(ctx, link)
	case strings.HasPrefix(data, "chal:"):
		b.sendChallengeExercises(ctx, link, strings.TrimPrefix(data, "chal:"))
	case strings.HasPrefix(data, "chex:"):
		b.startChallengeExercise(ctx, link, strings.TrimPrefix(data, "chex:"))
	case data == "rev":
		b.sendDueReviews(ctx, link)
	case data == "prog":
		b.sendProgress(ctx, link)
	case data == "end":
		b.endActiveSession(ctx, link)
	}
}
