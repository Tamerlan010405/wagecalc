package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	httpapi "lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/api/http"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/config"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/email"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/service"
	"lab.qoollo.com/practice/2025/wagecalc/wagecalc-backend/internal/storage/postgres"
)

type randomIDGenerator struct{}

func (randomIDGenerator) NewID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(bytes[:])
}

func openDB(ctx context.Context, databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	db, err := openDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected")
	calculationRepo := postgres.NewCalculationRepository(db, cfg.DBEncryptionKey)
	authRepo := postgres.NewAuthRepository(db)
	passwordResetSender, err := email.NewPasswordResetSender(
		email.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			Security: cfg.SMTPSecurity,
			Timeout:  10 * time.Second,
		},
	)
	if err != nil {
		logger.Error("email sender init failed", "error", err)
		os.Exit(1)
	}

	authService := service.NewAuthService(
		authRepo,
		passwordResetSender,
		cfg.PasswordResetSecret,
	)
	deletedSessions, err := authService.CleanupExpiredSessions(context.Background())
	if err != nil {
		logger.Warn("expired sessions cleanup failed", "error", err)
	} else {
		logger.Info("expired sessions cleanup completed", "deleted_sessions", deletedSessions)
	}
	calculationService := service.NewCalculationService(calculationRepo, randomIDGenerator{})
	router := httpapi.NewRouter(httpapi.RouterDeps{
		CalculationService: calculationService,
		AuthService:        authService,
		Logger:             logger,
		CookieSecure:       cfg.CookieSecure,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})

	logger.Info("server starting", "http_addr", cfg.HTTPAddr)

	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
