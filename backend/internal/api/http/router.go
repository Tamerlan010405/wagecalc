package httpapi

import (
	"log/slog"
	"net/http"
)

type CalculationService interface {
	CalculationWriteService
	CalculationReadService
}

type RouterDeps struct {
	CalculationService CalculationService
	AuthService        AuthService
	Logger             *slog.Logger
	CookieSecure       bool
	CORSAllowedOrigins []string
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	writeHandler := NewCalculationWriteHandler(deps.CalculationService)
	readHandler := NewCalculationReadHandler(deps.CalculationService)
	authHandler := NewAuthHandler(deps.AuthService, deps.CookieSecure)

	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/api/v1/calculations", RequireAuth(
		deps.AuthService,
		http.HandlerFunc(calculationsHandler(writeHandler, readHandler)),
	))
	mux.Handle("/api/v1/calculations/{id}", RequireAuth(
		deps.AuthService,
		http.HandlerFunc(calculationByIDHandler(readHandler)),
	))
	mux.HandleFunc("/", notFoundHandler)
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("/api/v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/v1/auth/me", authHandler.Me)
	mux.HandleFunc("/api/v1/auth/password-reset/request", authHandler.RequestPasswordReset)
	mux.HandleFunc("/api/v1/auth/password-reset/confirm", authHandler.ConfirmPasswordReset)
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}
	handler := JSONHeaders(mux)
	handler = Recovery(logger, handler)
	handler = RequestLogger(logger, handler)
	handler = RequestID(handler)
	handler = CORS(deps.CORSAllowedOrigins, handler)

	return handler
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func calculationsHandler(writeHandler *CalculationWriteHandler, readHandler *CalculationReadHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			writeHandler.Create(w, r)
		case http.MethodGet:
			readHandler.List(w, r)
		default:
			writeMethodNotAllowed(w, "GET, POST")
		}
	}
}

func calculationByIDHandler(readHandler *CalculationReadHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			readHandler.GetByID(w, r)
		case http.MethodDelete:
			readHandler.Delete(w, r)
		default:
			writeMethodNotAllowed(w, "GET, DELETE")
		}
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, newAPIError(http.StatusNotFound, codeNotFound, "not found", "", nil))
}
