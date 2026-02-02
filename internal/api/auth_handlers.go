package api

import (
	"net/http"
	"strings"

	"github.com/nf-software/nf-toolbox/internal/auth"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// HandleRegister creates a new user account
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Hash password
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user
	user, err := s.db.CreateUser(req.Email, hash)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, s.cfg.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Response
	resp := AuthResponse{Token: token}
	resp.User.ID = user.ID
	resp.User.Email = user.Email

	writeJSON(w, http.StatusCreated, resp)
}

// HandleLogin authenticates a user
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Get user
	user, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, s.cfg.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Response
	resp := AuthResponse{Token: token}
	resp.User.ID = user.ID
	resp.User.Email = user.Email

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetMe returns current user info
func (s *Server) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	user, err := s.db.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	tools, _ := s.db.GetUserTools(userID)
	sub, _ := s.db.GetSubscription(userID)

	resp := map[string]interface{}{
		"id":           user.ID,
		"email":        user.Email,
		"created_at":   user.CreatedAt,
		"tools":        tools,
		"subscription": sub,
	}

	writeJSON(w, http.StatusOK, resp)
}
