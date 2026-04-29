package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/httpx"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"github.com/AgataPalma/biblios/internal/users"
)

type Handler struct {
	userService *users.Service
	jwtSecret   string
	tokenStore  *tokenstore.Store
}

func NewHandler(userService *users.Service, jwtSecret string, tokenStore *tokenstore.Store) *Handler {
	return &Handler{userService: userService, jwtSecret: jwtSecret, tokenStore: tokenStore}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := users.RegisterInput{Email: req.Email, Username: req.Username, Password: req.Password}
	if err := users.ValidateRegisterInput(input); err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), input)
	if err != nil {
		switch err.Error() {
		case "email already registered":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		case "username already taken":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	token, jti, err := GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	if err = h.tokenStore.StoreToken(r.Context(), user.ID, jti, 24*time.Hour); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.Login(r.Context(), users.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, jti, err := GenerateToken(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	if err = h.tokenStore.StoreToken(r.Context(), user.ID, jti, 24*time.Hour); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.tokenStore.DeleteToken(r.Context(), claims.UserID, claims.ID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to logout")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
