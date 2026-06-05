package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"goshen/backend/internal/middleware"
	"goshen/backend/pkg/response"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "email and password required")
		return
	}
	admin, err := h.admins.FindByEmail(r.Context(), req.Email)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}
	claims := jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		response.InternalError(w)
		return
	}
	response.OK(w, map[string]string{"token": tokenStr})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value(middleware.AdminIDKey)
	response.OK(w, map[string]interface{}{"id": adminID})
}
