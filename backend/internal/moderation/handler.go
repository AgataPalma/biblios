package moderation

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPending(w http.ResponseWriter, r *http.Request) {
	var page int = 1
	var limit int = 20
	var err error

	if p := r.URL.Query().Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil || page < 1 {
			page = 1
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	var result ListSubmissionsResult
	result, err = h.service.ListPending(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list submissions")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	var id string = chi.URLParam(r, "id")

	var submission *books.Submission
	var err error
	submission, err = h.service.GetSubmission(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "submission not found")
		return
	}

	writeJSON(w, http.StatusOK, submission)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool
	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var id string = chi.URLParam(r, "id")

	var err error = h.service.Approve(r.Context(), ApproveInput{
		SubmissionID: id,
		ReviewerID:   claims.UserID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submission approved"})
}

type rejectRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool
	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var id string = chi.URLParam(r, "id")

	var req rejectRequest
	var err error = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	err = h.service.Reject(r.Context(), RejectInput{
		SubmissionID: id,
		ReviewerID:   claims.UserID,
		Reason:       req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submission rejected"})
}

type editAndApproveRequest struct {
	Title       string        `json:"title"`
	Description *string       `json:"description"`
	CoverURL    *string       `json:"cover_url"`
	Edition     books.Edition `json:"edition"`
}

func (h *Handler) EditAndApprove(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool
	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var id string = chi.URLParam(r, "id")

	var req editAndApproveRequest
	var err error = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}

	err = h.service.EditAndApprove(r.Context(), EditAndApproveInput{
		SubmissionID: id,
		ReviewerID:   claims.UserID,
		Title:        req.Title,
		Description:  req.Description,
		CoverURL:     req.CoverURL,
		Edition:      req.Edition,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "submission edited and approved"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
