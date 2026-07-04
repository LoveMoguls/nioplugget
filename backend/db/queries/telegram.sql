-- name: CreateTelegramLinkCode :one
INSERT INTO telegram_link_codes (code, student_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTelegramLinkCode :one
SELECT * FROM telegram_link_codes WHERE code = $1;

-- name: MarkTelegramLinkCodeUsed :exec
UPDATE telegram_link_codes SET used_at = NOW() WHERE code = $1;

-- name: UpsertTelegramLink :one
INSERT INTO telegram_links (student_id, telegram_user_id, chat_id)
VALUES ($1, $2, $3)
ON CONFLICT (student_id)
DO UPDATE SET telegram_user_id = $2, chat_id = $3, linked_at = NOW()
RETURNING *;

-- name: GetTelegramLinkByTelegramUserID :one
SELECT * FROM telegram_links WHERE telegram_user_id = $1;

-- name: ListTelegramLinks :many
SELECT * FROM telegram_links;

-- name: GetTelegramSession :one
SELECT * FROM telegram_sessions WHERE chat_id = $1;

-- name: UpsertTelegramSession :one
INSERT INTO telegram_sessions (chat_id, student_id, state, active_session_id, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (chat_id)
DO UPDATE SET student_id = $2, state = $3, active_session_id = $4, updated_at = NOW()
RETURNING *;

-- name: TryInsertTelegramReminder :execrows
INSERT INTO telegram_reminders (student_id, sent_on)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
