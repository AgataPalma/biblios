package library

import (
	"context"
	"encoding/json"
	"github.com/AgataPalma/biblios/internal/httpx"
	"net/http"
	"strconv"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/go-chi/chi/v5"
)

type ServiceInterface interface {
	GetMyLibrary(ctx context.Context, userID string, page, limit int) (ListLibraryResult, error)
	UpdateReadingStatus(ctx context.Context, copyID, userID string, input UpdateCopyInput) error
	RemoveCopy(ctx context.Context, copyID, userID string) error
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

type updateReadingStatusRequest struct {
	Status            string  `json:"status"`
	CurrentPage       *int    `json:"current_page"`
	StartedReadingAt  *string `json:"started_reading_at"`
	FinishedReadingAt *string `json:"finished_reading_at"`
	OwnedByUser       *bool   `json:"owned_by_user"`
	BorrowedFrom      *string `json:"borrowed_from"`
	Location          *string `json:"location"`
}

func (h *Handler) UpdateReadingStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")

	var req updateReadingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	valid := map[string]bool{"want_to_read": true, "reading": true, "read": true}
	if !valid[req.Status] {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "invalid status")
		return
	}

	input := UpdateCopyInput{
		Status:            req.Status,
		CurrentPage:       req.CurrentPage,
		StartedReadingAt:  req.StartedReadingAt,
		FinishedReadingAt: req.FinishedReadingAt,
		OwnedByUser:       req.OwnedByUser,
		BorrowedFrom:      req.BorrowedFrom,
		Location:          req.Location,
	}

	if err := h.service.UpdateReadingStatus(r.Context(), copyID, claims.UserID, input); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update copy")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "copy updated"})
}

func (h *Handler) RemoveCopy(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")

	if err := h.service.RemoveCopy(r.Context(), copyID, claims.UserID); err != nil {
		if err.Error() == "copy not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to remove copy")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "copy removed"})
}

func (h *Handler) GetMyLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	result, err := h.service.GetMyLibrary(r.Context(), claims.UserID, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get library")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}
