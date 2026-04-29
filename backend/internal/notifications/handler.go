package notifications

import (
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /notifications?type=&read=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := httpx.PaginationParams(r)
	q := r.URL.Query()

	notifs, total, err := h.service.List(r.Context(), claims.UserID, q.Get("type"), q.Get("read"), page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}
	if notifs == nil {
		notifs = []Notification{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": notifs,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// PUT /notifications/{id}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.service.MarkRead(r.Context(), id, claims.UserID); err != nil {
		if err.Error() == "notification not found or already read" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "notification marked as read"})
}

// PUT /notifications/read-all
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.MarkAllRead(r.Context(), claims.UserID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to mark notifications as read")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}
