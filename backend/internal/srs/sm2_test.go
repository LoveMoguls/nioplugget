package srs_test

import (
	"math"
	"testing"
	"time"

	"github.com/trollstaven/nioplugget/backend/internal/srs"
)

// fixed reference time for deterministic NextReview assertions
var now = time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)

func daysFromNow(d int) time.Time {
	return now.Add(time.Duration(d) * 24 * time.Hour)
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name           string
		input          srs.SM2Input
		wantInterval   int
		wantReps       int
		wantEFApprox   float64
		wantNextReview time.Time
	}{
		{
			name: "first review score 5",
			input: srs.SM2Input{
				Score:           5,
				EaseFactor:      2.5,
				IntervalDays:    0,
				RepetitionCount: 0,
			},
			wantInterval:   1,
			wantReps:       1,
			wantEFApprox:   2.6,
			wantNextReview: daysFromNow(1),
		},
		{
			name: "first review score 4",
			input: srs.SM2Input{
				Score:           4,
				EaseFactor:      2.5,
				IntervalDays:    0,
				RepetitionCount: 0,
			},
			wantInterval:   1,
			wantReps:       1,
			wantEFApprox:   2.5, // EF + (0.1 - 1*0.1) = EF + 0 = 2.5
			wantNextReview: daysFromNow(1),
		},
		{
			name: "first review score 3",
			input: srs.SM2Input{
				Score:           3,
				EaseFactor:      2.5,
				IntervalDays:    0,
				RepetitionCount: 0,
			},
			wantInterval:   1,
			wantReps:       1,
			wantEFApprox:   2.36, // EF + (0.1 - 2*0.12) = 2.5 + (0.1 - 0.24) = 2.5 - 0.14 = 2.36
			wantNextReview: daysFromNow(1),
		},
		{
			name: "second review score 5",
			input: srs.SM2Input{
				Score:           5,
				EaseFactor:      2.5,
				IntervalDays:    1,
				RepetitionCount: 1,
			},
			wantInterval:   6,
			wantReps:       2,
			wantEFApprox:   2.6,
			wantNextReview: daysFromNow(6),
		},
		{
			name: "third review score 5 with EF 2.5",
			input: srs.SM2Input{
				Score:           5,
				EaseFactor:      2.5,
				IntervalDays:    6,
				RepetitionCount: 2,
			},
			wantInterval:   15, // round(6*2.5) = 15
			wantReps:       3,
			wantEFApprox:   2.6,
			wantNextReview: daysFromNow(15),
		},
		{
			name: "score 2 fail resets",
			input: srs.SM2Input{
				Score:           2,
				EaseFactor:      2.5,
				IntervalDays:    6,
				RepetitionCount: 2,
			},
			wantInterval:   1,
			wantReps:       0,
			wantEFApprox:   2.18, // EF + (0.1 - 3*0.14) = 2.5 + (0.1 - 0.42) = 2.5 - 0.32 = 2.18
			wantNextReview: daysFromNow(1),
		},
		{
			name: "score 1 fail EF decreased but floored",
			input: srs.SM2Input{
				Score:           1,
				EaseFactor:      2.5,
				IntervalDays:    6,
				RepetitionCount: 2,
			},
			wantInterval:   1,
			wantReps:       0,
			wantEFApprox:   1.96, // EF + (0.1 - 4*0.18) = 2.5 + (0.1-0.72) = 2.5 - 0.62 = 1.88... wait let me recalc
			// formula: EF' = EF + (0.1 - (5-score) * (0.08 + (5-score)*0.02))
			// score=1: (5-1)=4, 0.08+4*0.02=0.08+0.08=0.16, 4*0.16=0.64, 0.1-0.64=-0.54
			// EF' = 2.5 - 0.54 = 1.96
			wantNextReview: daysFromNow(1),
		},
		{
			name: "EF at floor 1.3 score 1 stays at 1.3",
			input: srs.SM2Input{
				Score:           1,
				EaseFactor:      1.3,
				IntervalDays:    1,
				RepetitionCount: 1,
			},
			wantInterval:   1,
			wantReps:       0,
			wantEFApprox:   1.3, // would drop below floor, floored at 1.3
			wantNextReview: daysFromNow(1),
		},
		{
			name: "large interval 30 days score 5 grows correctly",
			input: srs.SM2Input{
				Score:           5,
				EaseFactor:      2.5,
				IntervalDays:    30,
				RepetitionCount: 5,
			},
			wantInterval:   75, // round(30*2.5) = 75
			wantReps:       6,
			wantEFApprox:   2.6,
			wantNextReview: daysFromNow(75),
		},
		{
			name: "default initial state first review works",
			input: srs.SM2Input{
				Score:           4,
				EaseFactor:      2.5,
				IntervalDays:    0,
				RepetitionCount: 0,
			},
			wantInterval:   1,
			wantReps:       1,
			wantEFApprox:   2.5,
			wantNextReview: daysFromNow(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := srs.Calculate(tt.input, now)

			if got.IntervalDays != tt.wantInterval {
				t.Errorf("IntervalDays = %d, want %d", got.IntervalDays, tt.wantInterval)
			}
			if got.RepetitionCount != tt.wantReps {
				t.Errorf("RepetitionCount = %d, want %d", got.RepetitionCount, tt.wantReps)
			}
			if !approxEqual(got.EaseFactor, tt.wantEFApprox) {
				t.Errorf("EaseFactor = %.4f, want %.4f", got.EaseFactor, tt.wantEFApprox)
			}
			if !got.NextReview.Equal(tt.wantNextReview) {
				t.Errorf("NextReview = %v, want %v", got.NextReview, tt.wantNextReview)
			}
		})
	}
}

func TestCalculateEFNeverBelowFloor(t *testing.T) {
	// Repeatedly fail to verify EF never drops below 1.3
	input := srs.SM2Input{
		Score:           1,
		EaseFactor:      2.5,
		IntervalDays:    0,
		RepetitionCount: 0,
	}

	for i := 0; i < 20; i++ {
		output := srs.Calculate(input, now)
		if output.EaseFactor < 1.3 {
			t.Errorf("iteration %d: EaseFactor %f dropped below 1.3 floor", i, output.EaseFactor)
		}
		input = srs.SM2Input{
			Score:           1,
			EaseFactor:      output.EaseFactor,
			IntervalDays:    output.IntervalDays,
			RepetitionCount: output.RepetitionCount,
		}
	}
}

func TestCalculateIntervalProgression(t *testing.T) {
	// Verify the 1 → 6 → n*EF progression with repeated successes
	input := srs.SM2Input{
		Score:           5,
		EaseFactor:      2.5,
		IntervalDays:    0,
		RepetitionCount: 0,
	}

	output1 := srs.Calculate(input, now)
	if output1.IntervalDays != 1 {
		t.Errorf("first review interval = %d, want 1", output1.IntervalDays)
	}

	input2 := srs.SM2Input{
		Score:           5,
		EaseFactor:      output1.EaseFactor,
		IntervalDays:    output1.IntervalDays,
		RepetitionCount: output1.RepetitionCount,
	}
	output2 := srs.Calculate(input2, now)
	if output2.IntervalDays != 6 {
		t.Errorf("second review interval = %d, want 6", output2.IntervalDays)
	}

	input3 := srs.SM2Input{
		Score:           5,
		EaseFactor:      output2.EaseFactor,
		IntervalDays:    output2.IntervalDays,
		RepetitionCount: output2.RepetitionCount,
	}
	output3 := srs.Calculate(input3, now)
	wantInterval3 := int(math.Round(float64(output2.IntervalDays) * output2.EaseFactor))
	if output3.IntervalDays != wantInterval3 {
		t.Errorf("third review interval = %d, want %d", output3.IntervalDays, wantInterval3)
	}
}
