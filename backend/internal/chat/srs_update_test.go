package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type fakeSRSStore struct {
	ChatStore // panics on anything else
	existing  *queries.ReviewSchedule
	upserted  *queries.UpsertReviewScheduleParams
}

func (f *fakeSRSStore) GetReviewSchedule(_ context.Context, _ queries.GetReviewScheduleParams) (queries.ReviewSchedule, error) {
	if f.existing == nil {
		return queries.ReviewSchedule{}, errors.New("not found")
	}
	return *f.existing, nil
}

func (f *fakeSRSStore) UpsertReviewSchedule(_ context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error) {
	f.upserted = &arg
	return queries.ReviewSchedule{}, nil
}

func TestUpdateReviewScheduleFirstReview(t *testing.T) {
	store := &fakeSRSStore{}
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	var student, exercise pgtype.UUID
	student.Scan("11111111-1111-1111-1111-111111111111")
	exercise.Scan("22222222-2222-2222-2222-222222222222")

	if err := UpdateReviewSchedule(context.Background(), store, student, exercise, 4, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.upserted == nil {
		t.Fatal("expected UpsertReviewSchedule to be called")
	}
	if store.upserted.RepetitionCount < 1 {
		t.Errorf("repetition count = %d, want >= 1", store.upserted.RepetitionCount)
	}
	if !store.upserted.NextReview.Time.After(now) {
		t.Errorf("next review %v not after now %v", store.upserted.NextReview.Time, now)
	}
}
