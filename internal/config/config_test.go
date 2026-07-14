package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		port     string // value forced into CHATER_PORT ("" means unset -> default)
		wantAddr string
	}{
		{name: "default when unset", port: "", wantAddr: ":8080"},
		{name: "reserved port", port: "8020", wantAddr: ":8020"},
		{name: "custom port", port: "9999", wantAddr: ":9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envPort, tt.port)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Addr != tt.wantAddr {
				t.Fatalf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
		})
	}
}
