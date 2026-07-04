package telegram

import (
	"strings"
	"testing"
)

func TestSplitMessageShort(t *testing.T) {
	parts := splitMessage("hej", 4096)
	if len(parts) != 1 || parts[0] != "hej" {
		t.Errorf("got %v", parts)
	}
}

func TestSplitMessageLongBreaksAtNewline(t *testing.T) {
	text := strings.Repeat("a", 50) + "\n" + strings.Repeat("b", 30)
	parts := splitMessage(text, 60)
	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2: %v", len(parts), parts)
	}
	if parts[0] != strings.Repeat("a", 50)+"\n" {
		t.Errorf("part 0 = %q", parts[0])
	}
	for _, p := range parts {
		if len([]rune(p)) > 60 {
			t.Errorf("part exceeds max: %d", len([]rune(p)))
		}
	}
}

func TestSplitMessageNonPositiveMax(t *testing.T) {
	for _, max := range []int{0, -1} {
		parts := splitMessage("hej", max)
		if len(parts) != 1 || parts[0] != "hej" {
			t.Errorf("max=%d: got %v", max, parts)
		}
	}
}

func TestSplitMessageNoBreakpoint(t *testing.T) {
	text := strings.Repeat("x", 130)
	parts := splitMessage(text, 60)
	if len(parts) != 3 {
		t.Fatalf("got %d parts, want 3", len(parts))
	}
}
