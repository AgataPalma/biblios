package auth

import (
	"encoding/json"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"net/http"
	"time"

	"github.com/AgataPalma/biblios/internal/users"
)

type Handler struct {
	userService *users.Service
	jwtSecret   string
	tokenStore  *tokenstore.Store
}

func NewHandler(userService *users.Service, jwtSecret string, tokenStore *tokenstore.Store) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
		tokenStore:  tokenStore,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponse struct {
	User users.User `json:"user"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var input users.RegisterInput = users.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// Validate input
	err = users.ValidateRegisterInput(input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Register user
	var user users.User
	user, err = h.userService.Register(r.Context(), input)
	if err != nil {
		if err.Error() == "email already registered" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{User: user})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var input users.LoginInput = users.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	var user users.User
	user, err = h.userService.Login(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var token string
	token, err = GenerateToken(user.ID, user.IsAdmin, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, User: user})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool

	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var remaining time.Duration = time.Until(claims.ExpiresAt.Time)
	var err error = h.tokenStore.RevokeToken(r.Context(), claims.ID, remaining)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool

	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":  claims.UserID,
		"is_admin": claims.IsAdmin,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
