// Package migrations embeds the goose SQL migration files so they travel inside
// the binary — the service applies them itself (no external migrate step needed
// for dev). sqlc reads the same files as its schema source.
package migrations

import "embed"

// FS holds every migration in this directory.
//
//go:embed *.sql
var FS embed.FS
