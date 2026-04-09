package auth

import (
	"encoding/json"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/httpx"
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

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  users.User `json:"user"`
}

/*type updateThemeRequest struct {
	Theme string `json:"theme"`
}*/

//----------------------------Functions---------------------------------------------//

func NewHandler(userService *users.Service, jwtSecret string, tokenStore *tokenstore.Store) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
		tokenStore:  tokenStore,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := users.RegisterInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	err = users.ValidateRegisterInput(input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), input)
	if err != nil {
		if err.Error() == "email already registered" {
			httpx.WriteError(w, http.StatusConflict, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create user")
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

	httpx.WriteJSON(w, http.StatusCreated, LoginResponse{Token: token, User: user})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	var err error

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := users.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.userService.Login(r.Context(), input)
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

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{Token: token, User: user})
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

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}
