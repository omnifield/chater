package httpapi

import "testing"

func TestHandleFromToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{name: "plain ascii unchanged", token: "alice", want: "alice"},
		{name: "percent-encoded cyrillic", token: "%D0%B5%D0%B3%D0%BE%D1%80", want: "егор"},
		{name: "encoded ascii round-trips", token: "%61%6C%69%63%65", want: "alice"},
		{name: "space", token: "a%20b", want: "a b"},
		{name: "malformed escape used verbatim", token: "%zz", want: "%zz"},
		{name: "no escapes verbatim", token: "bob.123_-", want: "bob.123_-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handleFromToken(tt.token); got != tt.want {
				t.Fatalf("handleFromToken(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}
