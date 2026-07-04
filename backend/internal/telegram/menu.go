package telegram

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

func (b *Bot) sendMainMenu(ctx context.Context, link queries.TelegramLink) {
	due, err := b.store.ListDueReviews(ctx, link.StudentID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list due reviews for menu")
	}
	kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{{Text: "📚 Plugga", CallbackData: "study"}},
		{{Text: fmt.Sprintf("⭐ Repetera (%d)", len(due)), CallbackData: "rev"}},
		{{Text: "🏆 Utmaningar", CallbackData: "chals"}},
		{{Text: "📊 Min progress", CallbackData: "prog"}},
	}}
	b.send(ctx, link.ChatID, "Vad vill du göra?", kb)
}

// Stubs — implemented in the browse/dialog task (Task 7).

func (b *Bot) sendSubjects(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendTopics(ctx context.Context, link queries.TelegramLink, subjectSlug string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendExercises(ctx context.Context, link queries.TelegramLink, subjectSlug, topicSlug string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendChallenges(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendChallengeExercises(ctx context.Context, link queries.TelegramLink, challengeID string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendDueReviews(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) sendProgress(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) startExercise(ctx context.Context, link queries.TelegramLink, exerciseID string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) startChallengeExercise(ctx context.Context, link queries.TelegramLink, exerciseID string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) endActiveSession(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}

func (b *Bot) handleDialogMessage(ctx context.Context, link queries.TelegramLink, tgSession queries.TelegramSession, text string) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}
