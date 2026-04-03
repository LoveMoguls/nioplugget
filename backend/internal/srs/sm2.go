// Package srs implements the SM-2 spaced repetition scheduling algorithm.
//
// SM-2 is a scheduling algorithm that determines optimal review intervals
// based on how well a student recalls material. Higher scores lead to
// longer intervals; failures reset to day 1.
package srs

import (
	"math"
	"time"
)

// SM2Input holds the current state of a card before a review.
type SM2Input struct {
	// Score is the quality of recall, 1-5 (1=complete failure, 5=perfect).
	Score int
	// EaseFactor is the current ease factor (default 2.5, floor 1.3).
	EaseFactor float64
	// IntervalDays is the number of days since the last review.
	IntervalDays int
	// RepetitionCount is how many times this card has been successfully recalled (score >= 3).
	RepetitionCount int
}

// SM2Output holds the updated scheduling state after a review.
type SM2Output struct {
	// EaseFactor is the updated ease factor after applying the SM-2 formula.
	EaseFactor float64
	// IntervalDays is the number of days until the next review.
	IntervalDays int
	// RepetitionCount is the updated count of consecutive successful recalls.
	RepetitionCount int
	// NextReview is the absolute time of the next scheduled review.
	NextReview time.Time
}

const (
	efFloor            = 1.3
	firstInterval      = 1
	secondInterval     = 6
)

// Calculate applies the SM-2 algorithm to produce the next review schedule.
//
// Interval rules:
//   - Score < 3 (fail): interval resets to 1 day, repetition count resets to 0
//   - Score >= 3, first success (reps=0): interval = 1 day
//   - Score >= 3, second success (reps=1): interval = 6 days
//   - Score >= 3, subsequent: interval = round(prev_interval * ease_factor)
//
// Ease factor formula: EF' = EF + (0.1 - (5-score) * (0.08 + (5-score)*0.02))
// The ease factor is floored at 1.3 and applied after interval calculation.
func Calculate(input SM2Input, now time.Time) SM2Output {
	var newInterval int
	var newReps int

	if input.Score < 3 {
		// Failed recall: reset interval and repetition count
		newInterval = firstInterval
		newReps = 0
	} else {
		// Successful recall: advance through interval schedule
		switch input.RepetitionCount {
		case 0:
			newInterval = firstInterval
		case 1:
			newInterval = secondInterval
		default:
			newInterval = int(math.Round(float64(input.IntervalDays) * input.EaseFactor))
		}
		newReps = input.RepetitionCount + 1
	}

	// Apply ease factor formula (always applied, even on failure)
	delta := float64(5 - input.Score)
	efDelta := 0.1 - delta*(0.08+delta*0.02)
	newEF := input.EaseFactor + efDelta

	// Enforce ease factor floor
	newEF = math.Max(newEF, efFloor)

	return SM2Output{
		EaseFactor:      newEF,
		IntervalDays:    newInterval,
		RepetitionCount: newReps,
		NextReview:      now.Add(time.Duration(newInterval) * 24 * time.Hour),
	}
}
