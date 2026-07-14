package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		port       string // value forced into CHATER_PORT ("" means unset -> default)
		dbPath     string // value forced into CHATER_DB_PATH ("" means unset -> default)
		wantAddr   string
		wantDBPath string
	}{
		{name: "defaults when unset", port: "", dbPath: "", wantAddr: ":8080", wantDBPath: "chater.db"},
		{name: "reserved port", port: "8020", dbPath: "", wantAddr: ":8020", wantDBPath: "chater.db"},
		{name: "custom port and db", port: "9999", dbPath: "/data/x.db", wantAddr: ":9999", wantDBPath: "/data/x.db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envPort, tt.port)
			t.Setenv(envDBPath, tt.dbPath)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Addr != tt.wantAddr {
				t.Fatalf("Addr = %q, want %q", cfg.Addr, tt.wantAddr)
			}
			if cfg.DBPath != tt.wantDBPath {
				t.Fatalf("DBPath = %q, want %q", cfg.DBPath, tt.wantDBPath)
			}
		})
	}
}
