-- name: CreateUser :one
INSERT INTO users (handle, created_at)
VALUES (?, ?)
RETURNING id, handle, created_at;

-- name: CreateRoom :one
INSERT INTO rooms (type, title, created_at)
VALUES (?, ?, ?)
RETURNING id, type, title, created_at;

-- name: AddParticipant :exec
INSERT INTO room_participants (room_id, user_id, role, joined_at)
VALUES (?, ?, ?, ?);

-- name: InsertMessage :one
INSERT INTO messages (room_id, author_id, body, created_at)
VALUES (?, ?, ?, ?)
RETURNING id, room_id, author_id, body, created_at;

-- ListMessagesByRoom returns one page of history, newest first.
-- Keyset pagination over (created_at, id): pass the last row seen as the cursor
-- to fetch the next (older) page. First page: cursor_created_at = '' (sentinel),
-- which disables the bound. The cursor param names repeat, so sqlc collapses
-- each to a single argument.
-- name: ListMessagesByRoom :many
SELECT id, room_id, author_id, body, created_at
FROM messages
WHERE room_id = @room_id
  AND (
      @cursor_created_at = ''
      OR created_at < @cursor_created_at
      OR (created_at = @cursor_created_at AND id < @cursor_id)
  )
ORDER BY created_at DESC, id DESC
LIMIT @page_limit;

-- ListRoomsForUser lists every room the user participates in, newest first.
-- name: ListRoomsForUser :many
SELECT r.id, r.type, r.title, r.created_at
FROM rooms r
JOIN room_participants p ON p.room_id = r.id
WHERE p.user_id = ?
ORDER BY r.created_at DESC, r.id DESC;
