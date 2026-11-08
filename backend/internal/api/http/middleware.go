package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
)

var defaultCORSAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:4173",
	"http://localhost:5173",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:4173",
	"http://127.0.0.1:5173",
}

func JSONHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic recovered",
					"request_id", RequestIDFromContext(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"error", recovered,
				)

				writeInternalError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	allowedOriginSet := corsOriginSet(allowedOrigins)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if _, ok := allowedOriginSet[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Request-ID")
		}

		if r.Method == http.MethodOptions {
			writeNoContent(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAuth(authService AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authService == nil {
			writeInternalError(w)
			return
		}
		token, ok := sessionTokenFromRequest(r)
		if !ok {
			writeError(w, unauthorizedError())
			return
		}
		user, err := authService.Me(r.Context(), token)
		if err != nil {
			writeError(w, mapAuthError(err))
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithAuthUser(r.Context(), user)))
	})
}

func corsOriginSet(origins []string) map[string]struct{} {
	if len(origins) == 0 {
		origins = defaultCORSAllowedOrigins
	}
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			set[origin] = struct{}{}
		}
	}
	return set
}
