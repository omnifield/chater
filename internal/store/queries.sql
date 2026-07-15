-- name: CreateUser :one
INSERT INTO users (handle, created_at)
VALUES (?, ?)
RETURNING id, handle, created_at;

-- name: GetUserByHandle :one
SELECT id, handle, created_at FROM users WHERE handle = ?;

-- name: GetRoom :one
SELECT id, type, title, created_at FROM rooms WHERE id = ?;

-- name: RoomParticipantExists :one
SELECT EXISTS(
    SELECT 1 FROM room_participants WHERE room_id = ? AND user_id = ?
) AS present;

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
SELECT m.id, m.room_id, m.author_id, m.body, m.created_at, u.handle AS author_handle
FROM messages m
JOIN users u ON u.id = m.author_id
WHERE m.room_id = @room_id
  AND (
      @cursor_created_at = ''
      OR m.created_at < @cursor_created_at
      OR (m.created_at = @cursor_created_at AND m.id < @cursor_id)
  )
ORDER BY m.created_at DESC, m.id DESC
LIMIT @page_limit;

-- ListRoomsForUser lists every room the user participates in, newest first.
-- name: ListRoomsForUser :many
SELECT r.id, r.type, r.title, r.created_at
FROM rooms r
JOIN room_participants p ON p.room_id = r.id
WHERE p.user_id = ?
ORDER BY r.created_at DESC, r.id DESC;
