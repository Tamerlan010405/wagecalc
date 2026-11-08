package config

import (
	"strings"
	"testing"
)

const (
	testEncryptionKey = "0123456789abcdef0123456789abcdef"
	testResetSecret   = "abcdef0123456789abcdef0123456789"
)

func TestLoadDBEncryptionKey(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		setValidEnv(t)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.DBEncryptionKey != testEncryptionKey {
			t.Fatal("Load() did not preserve DB_ENCRYPTION_KEY")
		}
	})

	tests := map[string]string{
		"missing": "",
		"short":   strings.Repeat("x", 31),
		"reused":  testResetSecret,
	}
	for name, key := range tests {
		t.Run(name, func(t *testing.T) {
			setValidEnv(t)
			t.Setenv("DB_ENCRYPTION_KEY", key)
			if _, err := Load(); err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://test")
	t.Setenv("DB_ENCRYPTION_KEY", testEncryptionKey)
	t.Setenv("PASSWORD_RESET_SECRET", testResetSecret)
	t.Setenv("SMTP_HOST", "localhost")
	t.Setenv("SMTP_FROM", "test@example.com")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
}
