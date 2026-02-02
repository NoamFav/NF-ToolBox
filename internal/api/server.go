package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nf-software/nf-toolbox/internal/auth"
	"github.com/nf-software/nf-toolbox/internal/config"
	"github.com/nf-software/nf-toolbox/internal/db"
)

type Server struct {
	db  *db.DB
	cfg *config.Config
}

func NewServer(database *db.DB, cfg *config.Config) *Server {
	return &Server{
		db:  database,
		cfg: cfg,
	}
}

// Context key for user ID
type contextKey string

const userIDKey contextKey = "userID"

// AuthMiddleware validates JWT and adds user ID to context
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Extract bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		// Validate token
		claims, err := auth.ValidateToken(parts[1], s.cfg.JWTSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserID extracts user ID from context
func getUserID(r *http.Request) string {
	if id, ok := r.Context().Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// JSON helpers

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
