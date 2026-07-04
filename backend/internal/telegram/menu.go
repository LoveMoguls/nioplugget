package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/challenges"
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

func (b *Bot) sendSubjects(ctx context.Context, link queries.TelegramLink) {
	subjects, err := b.store.ListSubjects(ctx)
	if err != nil || len(subjects) == 0 {
		b.send(ctx, link.ChatID, "Kunde inte hämta ämnen just nu.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, s := range subjects {
		rows = append(rows, []InlineKeyboardButton{{Text: s.Name, CallbackData: "subj:" + s.Slug}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Välj ämne:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendTopics(ctx context.Context, link queries.TelegramLink, subjectSlug string) {
	subject, err := b.store.GetSubjectBySlug(ctx, subjectSlug)
	if err != nil {
		b.send(ctx, link.ChatID, "Ämnet hittades inte.", nil)
		return
	}
	topics, err := b.store.ListTopicsBySubjectID(ctx, subject.ID)
	if err != nil || len(topics) == 0 {
		b.send(ctx, link.ChatID, "Inga områden hittades.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, t := range topics {
		rows = append(rows, []InlineKeyboardButton{{Text: t.Name, CallbackData: "top:" + subjectSlug + ":" + t.Slug}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Ämnen", CallbackData: "study"}})
	b.send(ctx, link.ChatID, "Välj område:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendExercises(ctx context.Context, link queries.TelegramLink, subjectSlug, topicSlug string) {
	topic, err := b.store.GetTopicBySlug(ctx, queries.GetTopicBySlugParams{SubjectSlug: subjectSlug, TopicSlug: topicSlug})
	if err != nil {
		b.send(ctx, link.ChatID, "Området hittades inte.", nil)
		return
	}
	exercises, err := b.store.ListExercisesByTopicID(ctx, topic.ID)
	if err != nil || len(exercises) == 0 {
		b.send(ctx, link.ChatID, "Inga övningar hittades.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, e := range exercises {
		rows = append(rows, []InlineKeyboardButton{{Text: e.Title, CallbackData: "ex:" + uuidToString(e.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Områden", CallbackData: "subj:" + subjectSlug}})
	b.send(ctx, link.ChatID, "Välj övning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendChallenges(ctx context.Context, link queries.TelegramLink) {
	student, err := b.store.GetStudentByID(ctx, link.StudentID)
	if err != nil {
		b.send(ctx, link.ChatID, "Något gick fel.", nil)
		return
	}
	chals, err := b.store.ListPublishedChallengesByParentID(ctx, student.ParentID)
	if err != nil || len(chals) == 0 {
		b.send(ctx, link.ChatID, "Inga utmaningar ännu — be din förälder ladda upp en läxa på webben!", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, c := range chals {
		rows = append(rows, []InlineKeyboardButton{{Text: c.CoverEmoji + " " + c.Title, CallbackData: "chal:" + uuidToString(c.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Välj utmaning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendChallengeExercises(ctx context.Context, link queries.TelegramLink, challengeIDStr string) {
	exercises, err := b.store.ListChallengeExercisesWithProgress(ctx, queries.ListChallengeExercisesWithProgressParams{
		ChallengeID: parseUUID(challengeIDStr),
		StudentID:   link.StudentID,
	})
	if err != nil || len(exercises) == 0 {
		b.send(ctx, link.ChatID, "Inga övningar i den här utmaningen.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, e := range exercises {
		label := e.Title
		if e.Score.Valid {
			label = strings.Repeat("⭐", challenges.ScoreToStars(int(e.Score.Int32))) + " " + label
		}
		rows = append(rows, []InlineKeyboardButton{{Text: label, CallbackData: "chex:" + uuidToString(e.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Utmaningar", CallbackData: "chals"}})
	b.send(ctx, link.ChatID, "Välj övning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendDueReviews(ctx context.Context, link queries.TelegramLink) {
	due, err := b.store.ListDueReviews(ctx, link.StudentID)
	if err != nil {
		b.send(ctx, link.ChatID, "Kunde inte hämta repetitioner.", nil)
		return
	}
	if len(due) == 0 {
		b.send(ctx, link.ChatID, "Inget att repetera just nu — snyggt jobbat! 🎉", nil)
		b.sendMainMenu(ctx, link)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, r := range due {
		rows = append(rows, []InlineKeyboardButton{{Text: r.SubjectName + ": " + r.ExerciseTitle, CallbackData: "ex:" + uuidToString(r.ExerciseID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Dags att repetera:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendProgress(ctx context.Context, link queries.TelegramLink) {
	rows, err := b.store.GetStudentProgressBySubject(ctx, link.StudentID)
	if err != nil || len(rows) == 0 {
		b.send(ctx, link.ChatID, "Ingen progress ännu — kör din första övning! 💪", nil)
		return
	}
	var sb strings.Builder
	sb.WriteString("📊 Din progress:\n\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "%s: %d pass, snitt %.1f/5, %d övningar\n", r.Name, r.TotalSessions, r.AvgScore, r.UniqueExercises)
	}
	kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{{Text: "⬅️ Meny", CallbackData: "menu"}}}}
	b.send(ctx, link.ChatID, sb.String(), kb)
}
