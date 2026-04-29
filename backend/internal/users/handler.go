package users

import (
	"encoding/json"
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.service.GetByID(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Username  *string `json:"username"`
		Bio       *string `json:"bio"`
		AvatarURL *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == nil && req.Bio == nil && req.AvatarURL == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "at least one field is required")
		return
	}
	user, err := h.service.UpdateProfile(r.Context(), claims.UserID, req.Username, req.Bio, req.AvatarURL)
	if err != nil {
		if err.Error() == "username already taken" {
			httpx.WriteError(w, http.StatusConflict, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.CurrentPassword == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "email and current_password are required")
		return
	}
	if err := h.service.UpdateEmail(r.Context(), claims.UserID, req.CurrentPassword, req.Email); err != nil {
		switch err.Error() {
		case "current password is incorrect":
			httpx.WriteError(w, http.StatusUnauthorized, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update email")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "email updated"})
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "current_password and new_password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "new password must be at least 8 characters")
		return
	}
	if err := h.service.UpdatePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		switch err.Error() {
		case "current password is incorrect":
			httpx.WriteError(w, http.StatusUnauthorized, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update password")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func (h *Handler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Theme == "" {
		httpx.WriteError(w, http.StatusBadRequest, "theme is required")
		return
	}
	if err := h.service.UpdateTheme(r.Context(), claims.UserID, req.Theme); err != nil {
		if err.Error() == "invalid theme" {
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update theme")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "theme updated"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.service.DeleteUser(r.Context(), claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
}
