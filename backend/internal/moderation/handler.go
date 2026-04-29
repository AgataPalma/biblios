package moderation

import (
	"encoding/json"
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/AgataPalma/biblios/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /moderation/submissions
func (h *Handler) ListPending(w http.ResponseWriter, r *http.Request) {
	page, limit := httpx.PaginationParams(r)
	subs, total, err := h.service.ListPending(r.Context(), page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list submissions")
		return
	}
	if subs == nil {
		subs = []books.Submission{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"submissions": subs,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// GET /moderation/submissions/{id}
func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sub, err := h.service.GetSubmission(r.Context(), id)
	if err != nil {
		if err.Error() == "submission not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get submission")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sub)
}

// PUT /moderation/submissions/{id}/approve
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	if err := h.service.Approve(r.Context(), id, claims.UserID); err != nil {
		switch err.Error() {
		case "submission not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "submission is not pending":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to approve submission")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "submission approved"})
}

// PUT /moderation/submissions/{id}/reject
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Reject(r.Context(), id, claims.UserID, req.Reason); err != nil {
		switch err.Error() {
		case "submission not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "submission is not pending":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		case "rejection reason is required":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to reject submission")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "submission rejected"})
}

// PUT /moderation/submissions/{id}/edit
func (h *Handler) EditAndApprove(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	// Accept optional book/edition update fields
	var req struct {
		Title          *string             `json:"title"`
		Description    *string             `json:"description"`
		Authors        []string            `json:"authors"`
		Genres         []string            `json:"genres"`
		SeriesName     *string             `json:"series_name"`
		SeriesPosition *float64            `json:"series_position"`
		EditionID      string              `json:"edition_id"`
		Edition        *books.EditionInput `json:"edition"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var bookInput *books.UpdateBookInput
	if req.Title != nil || req.Description != nil || len(req.Authors) > 0 ||
		len(req.Genres) > 0 || req.Edition != nil {
		bookInput = &books.UpdateBookInput{
			Title:          req.Title,
			Description:    req.Description,
			Authors:        req.Authors,
			Genres:         req.Genres,
			SeriesName:     req.SeriesName,
			SeriesPosition: req.SeriesPosition,
			EditionID:      req.EditionID,
			Edition:        req.Edition,
		}
	}

	if err := h.service.EditAndApprove(r.Context(), id, claims.UserID, bookInput); err != nil {
		switch err.Error() {
		case "submission not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "submission is not pending":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to edit and approve submission")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "submission edited and approved"})
}

// GET /moderation/logs
func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	page, limit := httpx.PaginationParams(r)
	logs, total, err := h.service.ListLogs(r.Context(), page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list moderation logs")
		return
	}
	if logs == nil {
		logs = []ModerationLog{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
