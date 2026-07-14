-- +goose Up
-- +goose StatementBegin

-- Minimal identity: id + unique handle. No auth providers in v0 (token-stub later).
CREATE TABLE users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    handle     TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL          -- ISO-8601 UTC, set by the application
);

-- A dialog and a group are the SAME entity; the difference is `type` plus the
-- number of participants. `title` is nullable (dialogs typically have none).
CREATE TABLE rooms (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    type       TEXT NOT NULL CHECK (type IN ('dialog', 'group')),
    title      TEXT,
    created_at TEXT NOT NULL
);

CREATE TABLE room_participants (
    room_id   INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      TEXT,                   -- nullable; role semantics are a later concern
    joined_at TEXT NOT NULL,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id    INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    author_id  INTEGER NOT NULL REFERENCES users(id),
    body       TEXT NOT NULL,
    created_at TEXT NOT NULL
);

-- History pagination is keyset over (room_id, created_at, id); this index backs it.
CREATE INDEX idx_messages_room_created ON messages (room_id, created_at, id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_room_created;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS room_participants;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
