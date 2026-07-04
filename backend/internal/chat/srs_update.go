package chat

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
	"github.com/trollstaven/nioplugget/backend/internal/srs"
)

// UpdateReviewSchedule recalculates the SM-2 review schedule after a scored
// session. Shared by the HTTP EndSession handler and the Telegram bot.
func UpdateReviewSchedule(ctx context.Context, store ChatStore, studentID, exerciseID pgtype.UUID, score int, now time.Time) error {
	input := srs.SM2Input{
		Score:           score,
		EaseFactor:      2.5, // SM-2 default
		IntervalDays:    1,
		RepetitionCount: 0,
	}
	existing, err := store.GetReviewSchedule(ctx, queries.GetReviewScheduleParams{
		StudentID:  studentID,
		ExerciseID: exerciseID,
	})
	if err == nil {
		input.EaseFactor = float64(existing.EaseFactor)
		input.IntervalDays = int(existing.IntervalDays)
		input.RepetitionCount = int(existing.RepetitionCount)
	}
	output := srs.Calculate(input, now)
	_, err = store.UpsertReviewSchedule(ctx, queries.UpsertReviewScheduleParams{
		StudentID:       studentID,
		ExerciseID:      exerciseID,
		EaseFactor:      float32(output.EaseFactor),
		IntervalDays:    int32(output.IntervalDays),
		RepetitionCount: int32(output.RepetitionCount),
		NextReview:      pgtype.Timestamptz{Time: output.NextReview, Valid: true},
	})
	return err
}
