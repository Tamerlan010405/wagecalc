package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

const defaultHTTPAddr = ":8080"

type Config struct {
	HTTPAddr            string
	DatabaseURL         string
	DBEncryptionKey     string
	CookieSecure        bool
	CORSAllowedOrigins  []string
	PasswordResetSecret string
	SMTPHost            string
	SMTPPort            int
	SMTPUsername        string
	SMTPPassword        string
	SMTPFrom            string
	SMTPSecurity        string
}

func Load() (Config, error) {

	httpAddr := strings.TrimSpace(os.Getenv("HTTP_ADDR"))
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	dbEncryptionKey := strings.TrimSpace(os.Getenv("DB_ENCRYPTION_KEY"))
	if len([]byte(dbEncryptionKey)) < 32 {
		return Config{}, errors.New("DB_ENCRYPTION_KEY must contain at least 32 bytes")
	}
	cookieSecure := false
	if value := strings.TrimSpace(os.Getenv("COOKIE_SECURE")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, errors.New("COOKIE_SECURE must be a boolean")
		}
		cookieSecure = parsed
	}
	corsAllowedOrigins, err := parseCORSAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		return Config{}, err
	}
	passwordResetSecret := strings.TrimSpace(
		os.Getenv("PASSWORD_RESET_SECRET"),
	)
	if len(passwordResetSecret) < 32 {
		return Config{}, errors.New(
			"PASSWORD_RESET_SECRET must contain at least 32 characters",
		)
	}
	if dbEncryptionKey == passwordResetSecret {
		return Config{}, errors.New("DB_ENCRYPTION_KEY must differ from PASSWORD_RESET_SECRET")
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if smtpHost == "" {
		return Config{}, errors.New("SMTP_HOST is required")
	}

	smtpPort := 587
	if value := strings.TrimSpace(os.Getenv("SMTP_PORT")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 65535 {
			return Config{}, errors.New("invalid SMTP_PORT")
		}
		smtpPort = parsed
	}

	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if smtpFrom == "" {
		return Config{}, errors.New("SMTP_FROM is required")
	}

	smtpSecurity := strings.ToLower(
		strings.TrimSpace(os.Getenv("SMTP_SECURITY")),
	)
	if smtpSecurity == "" {
		smtpSecurity = "starttls"
	}
	if smtpSecurity != "none" && smtpSecurity != "starttls" && smtpSecurity != "tls" {
		return Config{}, errors.New(
			"SMTP_SECURITY must be one of: none, starttls, tls",
		)
	}

	smtpUsername := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if (smtpUsername == "") != (smtpPassword == "") {
		return Config{}, errors.New(
			"SMTP_USERNAME and SMTP_PASSWORD must be set together",
		)
	}
	return Config{
		HTTPAddr:            httpAddr,
		DatabaseURL:         databaseURL,
		DBEncryptionKey:     dbEncryptionKey,
		CookieSecure:        cookieSecure,
		CORSAllowedOrigins:  corsAllowedOrigins,
		PasswordResetSecret: passwordResetSecret,
		SMTPHost:            smtpHost,
		SMTPPort:            smtpPort,
		SMTPUsername:        smtpUsername,
		SMTPPassword:        smtpPassword,
		SMTPFrom:            smtpFrom,
		SMTPSecurity:        smtpSecurity,
	}, nil
}

func parseCORSAllowedOrigins(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		if origin == "*" {
			return nil, errors.New("CORS_ALLOWED_ORIGINS cannot contain * when credentials are enabled")
		}
		if _, ok := seen[origin]; ok {
			continue
		}

		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	return origins, nil
}
