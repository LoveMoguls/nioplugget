package telegram

import (
	"context"
	"strings"
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

func TestGenerateLinkCode(t *testing.T) {
	code, err := generateLinkCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 8 {
		t.Errorf("len = %d, want 8", len(code))
	}
	for _, c := range code {
		if !strings.ContainsRune(codeAlphabet, c) {
			t.Errorf("invalid char %q", c)
		}
	}
	code2, _ := generateLinkCode()
	if code == code2 {
		t.Error("two codes should differ")
	}
}

type fakeLinkStore struct {
	Store
	created *queries.CreateTelegramLinkCodeParams
}

func (f *fakeLinkStore) CreateTelegramLinkCode(_ context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error) {
	f.created = &arg
	return queries.TelegramLinkCode{Code: arg.Code, StudentID: arg.StudentID, ExpiresAt: arg.ExpiresAt}, nil
}

func TestCreateLinkCode(t *testing.T) {
	store := &fakeLinkStore{}
	studentID := "660e8400-e29b-41d4-a716-446655440000"

	code, link, err := createLinkCode(context.Background(), store, "nioplugget_bot", studentID)
	if err != nil {
		t.Fatal(err)
	}

	if len(code) != 8 {
		t.Errorf("code len = %d, want 8", len(code))
	}
	wantLink := "https://t.me/nioplugget_bot?start=" + code
	if link != wantLink {
		t.Errorf("link = %q, want %q", link, wantLink)
	}

	if store.created == nil {
		t.Fatal("expected CreateTelegramLinkCode to be called")
	}
	if store.created.Code != code {
		t.Errorf("stored code = %q, want %q", store.created.Code, code)
	}
	if uuidToString(store.created.StudentID) != studentID {
		t.Errorf("stored studentID = %q, want %q", uuidToString(store.created.StudentID), studentID)
	}
	if !store.created.ExpiresAt.Valid {
		t.Error("expected ExpiresAt to be valid")
	}
}

type erroringLinkStore struct {
	Store
}

func (erroringLinkStore) CreateTelegramLinkCode(_ context.Context, _ queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error) {
	return queries.TelegramLinkCode{}, context.DeadlineExceeded
}

func TestCreateLinkCode_StoreError(t *testing.T) {
	_, _, err := createLinkCode(context.Background(), erroringLinkStore{}, "nioplugget_bot", "660e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Fatal("expected error from store")
	}
}
