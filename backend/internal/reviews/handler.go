package reviews

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /books/{id}/reviews
func (h *Handler) GetBookReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")

	page, limit := 1, 10
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

	result, err := h.service.GetBookReviews(r.Context(), bookID, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get reviews")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GET /books/{id}/reviews/me
func (h *Handler) GetMyReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookID := chi.URLParam(r, "id")

	review, err := h.service.GetMyReview(r.Context(), bookID, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get review")
		return
	}
	if review == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"review": nil})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"review": review})
}

type upsertReviewRequest struct {
	Rating   int     `json:"rating"`
	Body     *string `json:"body"`
	IsPublic *bool   `json:"is_public"`
}

// POST /books/{id}/reviews
func (h *Handler) UpsertReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookID := chi.URLParam(r, "id")

	var req upsertReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		writeError(w, http.StatusUnprocessableEntity, "rating must be between 1 and 5")
		return
	}

	review, err := h.service.UpsertReview(r.Context(), UpsertReviewInput{
		BookID:   bookID,
		UserID:   claims.UserID,
		Rating:   req.Rating,
		Body:     req.Body,
		IsPublic: req.IsPublic,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save review")
		return
	}

	writeJSON(w, http.StatusOK, review)
}

// PUT /books/{id}/reviews/me  (alias — same as POST for upsert semantics)
func (h *Handler) UpdateMyReview(w http.ResponseWriter, r *http.Request) {
	h.UpsertReview(w, r)
}

// DELETE /books/{id}/reviews/me
func (h *Handler) DeleteMyReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookID := chi.URLParam(r, "id")

	if err := h.service.DeleteReview(r.Context(), bookID, claims.UserID); err != nil {
		if err.Error() == "review not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete review")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "review deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
