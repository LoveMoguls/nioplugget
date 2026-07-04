package telegram

import (
	"context"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Sender is the subset of the Telegram API the bot sends through (mockable).
type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error
	SendChatAction(ctx context.Context, chatID int64, action string) error
	AnswerCallbackQuery(ctx context.Context, id string) error
}

// CompleteFunc produces a full AI reply (injectable for tests).
type CompleteFunc func(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam) (string, error)

// ScoreFunc scores a finished session (injectable for tests).
type ScoreFunc func(ctx context.Context, apiKey, exerciseTitle string, messages []queries.Message) (*chat.ScoreResult, error)

type Bot struct {
	api      Sender
	store    Store
	encSvc   *apikey.EncryptionService
	complete CompleteFunc
	score    ScoreFunc
}

func NewBot(api Sender, store Store, encSvc *apikey.EncryptionService) *Bot {
	return &Bot{
		api:    api,
		store:  store,
		encSvc: encSvc,
		complete: func(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam) (string, error) {
			return chat.CompleteChatResponse(ctx, apiKey, systemPrompt, messages)
		},
		score: chat.ScoreSession,
	}
}

// recoverPanic logs a panic instead of letting it kill the process.
func recoverPanic(where string) {
	if r := recover(); r != nil {
		log.Error().Interface("panic", r).Str("where", where).Msg("telegram: recovered panic")
	}
}

// Run long-polls Telegram until ctx is canceled.
func Run(ctx context.Context, api *API, bot *Bot) {
	var offset int64
	backoff := time.Second
	for ctx.Err() == nil {
		updates, err := api.GetUpdates(ctx, offset, 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error().Err(err).Msg("telegram getUpdates failed")
			time.Sleep(backoff)
			if backoff < time.Minute {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
		for _, upd := range updates {
			offset = upd.UpdateID + 1
			func() {
				defer recoverPanic("handle update")
				bot.HandleUpdate(ctx, upd)
			}()
		}
	}
}

// send splits long texts and attaches the keyboard to the last part.
func (b *Bot) send(ctx context.Context, chatID int64, text string, kb *InlineKeyboardMarkup) {
	parts := splitMessage(text, telegramMessageLimit)
	for i, part := range parts {
		var partKB *InlineKeyboardMarkup
		if i == len(parts)-1 {
			partKB = kb
		}
		if err := b.api.SendMessage(ctx, chatID, part, partKB); err != nil {
			log.Error().Err(err).Int64("chat_id", chatID).Msg("telegram send failed")
		}
	}
}
