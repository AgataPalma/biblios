package auth

import (
	"encoding/json"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"net/http"
	"time"

	"github.com/AgataPalma/biblios/internal/users"
)

//-------------------Types------------------------

type Handler struct {
	userService *users.Service
	jwtSecret   string
	tokenStore  *tokenstore.Store
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

type updateThemeRequest struct {
	Theme string `json:"theme"`
}

//----------------------------Functions---------------------------------------------//

func NewHandler(userService *users.Service, jwtSecret string, tokenStore *tokenstore.Store) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
		tokenStore:  tokenStore,
	}
}

// POST /api/v1/auth/register

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := users.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	err = users.ValidateRegisterInput(input)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), input)
	if err != nil {
		if err.Error() == "email already registered" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, jti, err := GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	if err = h.tokenStore.StoreToken(r.Context(), user.ID, jti, 24*time.Hour); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	writeJSON(w, http.StatusCreated, loginResponse{Token: token, User: user})
}

// POST /api/v1/auth/login

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := users.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.userService.Login(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, jti, err := GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	if err = h.tokenStore.StoreToken(r.Context(), user.ID, jti, 24*time.Hour); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, User: user})
}

// POST /api/v1/auth/logout

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.tokenStore.DeleteToken(r.Context(), claims.UserID, claims.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
