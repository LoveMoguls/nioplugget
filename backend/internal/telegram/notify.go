package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// RunReminderLoop checks hourly for due reviews and sends at most one
// reminder per student per day, inside the 15:00–20:00 local window.
func RunReminderLoop(ctx context.Context, bot *Bot) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bot.sendDueReminders(ctx, time.Now())
		}
	}
}

func (b *Bot) sendDueReminders(ctx context.Context, now time.Time) {
	if now.Hour() < 15 || now.Hour() >= 20 {
		return
	}
	links, err := b.store.ListTelegramLinks(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list telegram links")
		return
	}
	for _, link := range links {
		due, err := b.store.ListDueReviews(ctx, link.StudentID)
		if err != nil || len(due) == 0 {
			continue
		}
		inserted, err := b.store.TryInsertTelegramReminder(ctx, queries.TryInsertTelegramReminderParams{
			StudentID: link.StudentID,
			SentOn:    pgtype.Date{Time: now, Valid: true},
		})
		if err != nil || inserted == 0 {
			continue // already reminded today
		}
		kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "⭐ Kör nu", CallbackData: "rev"}},
		}}
		b.send(ctx, link.ChatID, fmt.Sprintf("Dags att repetera! 🧠 Du har %d övning(ar) som väntar, t.ex. %s.", len(due), due[0].ExerciseTitle), kb)
	}
}

// ChallengePublished notifies all linked students in the challenge's family.
// Implements challenges.Notifier.
func (b *Bot) ChallengePublished(ctx context.Context, challengeID pgtype.UUID) {
	challenge, err := b.store.GetChallengeByID(ctx, challengeID)
	if err != nil || !challenge.Published {
		return
	}
	links, err := b.store.ListTelegramLinks(ctx)
	if err != nil {
		return
	}
	for _, link := range links {
		student, err := b.store.GetStudentByID(ctx, link.StudentID)
		if err != nil || student.ParentID != challenge.ParentID {
			continue
		}
		kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🏆 Öppna", CallbackData: "chal:" + uuidToString(challenge.ID)}},
		}}
		b.send(ctx, link.ChatID, fmt.Sprintf("Ny läxutmaning: %s %s", challenge.CoverEmoji, challenge.Title), kb)
	}
}
