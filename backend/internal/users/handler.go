package users

import (
	"encoding/json"
	"fmt"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/httpx"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Me GET /api/v1/auth/me
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

// UpdateUser PUT /api/v1/users/me
// Accepts any combination of: username, bio, avatar_url.
// At least one field must be provided. All fields are optional and independent —
// omitting a field (or sending null) leaves it unchanged via COALESCE in SQL.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Username  *string `json:"username"`
		Bio       *string `json:"bio"`
		AvatarUrl *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Reject requests that send nothing at all
	if req.Username == nil && req.Bio == nil && req.AvatarUrl == nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "at least one field is required")
		return
	}

	// Validate bio length if provided
	if req.Bio != nil && len(*req.Bio) > 500 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "bio must be 500 characters or fewer")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), UpdateUserInput{
		UserID:    claims.UserID,
		Username:  req.Username,
		Bio:       req.Bio,
		AvatarUrl: req.AvatarUrl,
	})

	if err != nil {
		switch err.Error() {
		case "email already in use", "username already in use":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update profile")
		}
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
		NewEmail        string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentPassword == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "current_password is required")
		return
	}

	if isEmailValid(req.NewEmail) == false {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "new_email is not valid")
		return
	}

	err := h.service.UpdateEmail(r.Context(), UpdateEmailInput{
		UserID:          claims.UserID,
		CurrentPassword: req.CurrentPassword,
	})
	if err != nil {
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

// UpdatePassword PUT /api/v1/users/me/password
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

	err := h.service.UpdatePassword(r.Context(), UpdatePasswordInput{
		UserID:          claims.UserID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
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

// DeleteUser DELETE /api/v1/users/me
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

// UpdateTheme PUT /api/v1/users/me/theme
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
		if err.Error() == fmt.Sprintf("invalid theme: %s", req.Theme) {
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update theme")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "theme updated"})
}
