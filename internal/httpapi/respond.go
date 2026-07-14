package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/omnifield/chater/internal/store"
)

// maxBodyBytes caps request bodies to keep a stray/huge payload from allocating
// unbounded memory.
const maxBodyBytes = 1 << 20 // 1 MiB

func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn("encode response", "err", err)
	}
}

type errorBody struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, logger *slog.Logger, status int, msg string) {
	writeJSON(w, logger, status, errorBody{Error: msg})
}

// decodeJSON reads a size-capped JSON body into dst, writing the appropriate
// 4xx and returning false on malformed or oversized input.
func decodeJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			writeError(w, logger, http.StatusRequestEntityTooLarge, "request body too large")
			return false
		}
		writeError(w, logger, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}

// pathID parses the {id} path segment as a positive int64.
func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// --- opaque history cursor: base64url of "created_at\x00id" ---

func encodeCursor(c store.Cursor) string {
	raw := c.CreatedAt + "\x00" + strconv.FormatInt(c.ID, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(s string) (store.Cursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return store.Cursor{}, fmt.Errorf("cursor: bad encoding")
	}
	parts := strings.SplitN(string(b), "\x00", 2)
	if len(parts) != 2 {
		return store.Cursor{}, fmt.Errorf("cursor: bad format")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return store.Cursor{}, fmt.Errorf("cursor: bad id")
	}
	return store.Cursor{CreatedAt: parts[0], ID: id}, nil
}
