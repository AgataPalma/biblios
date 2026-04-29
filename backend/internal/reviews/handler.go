package reviews

import (
	"encoding/json"
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

// GET /books/{id}/reviews
func (h *Handler) ListPublicReviews(w http.ResponseWriter, r *http.Request) {
	bookID := chi.URLParam(r, "id")
	page, limit := httpx.PaginationParams(r)

	// Caller ID for liked_by_me flag — optional (public endpoint)
	callerID := ""
	if claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims); ok {
		callerID = claims.UserID
	}

	resp, err := h.service.ListPublicReviews(r.Context(), bookID, callerID, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list reviews")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

// GET /books/{id}/reviews/me
func (h *Handler) GetMyReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	bookID := chi.URLParam(r, "id")

	rev, err := h.service.GetMyReview(r.Context(), bookID, claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get review")
		return
	}
	if rev == nil {
		httpx.WriteJSON(w, http.StatusOK, nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rev)
}

// POST /books/{id}/reviews  — creates or updates the caller's review
func (h *Handler) UpsertReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	bookID := chi.URLParam(r, "id")

	var req struct {
		Rating   float64 `json:"rating"`
		Body     *string `json:"body"`
		IsPublic *bool   `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	rev, err := h.service.UpsertReview(r.Context(), UpsertReviewInput{
		BookID:   bookID,
		UserID:   claims.UserID,
		Rating:   req.Rating,
		Body:     req.Body,
		IsPublic: isPublic,
	})
	if err != nil {
		switch err.Error() {
		case "rating must be between 0.0 and 5.0",
			"rating must be in 0.1 increments",
			"review body must be 5000 characters or fewer":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to save review")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, rev)
}

// DELETE /books/{id}/reviews/me
func (h *Handler) DeleteMyReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	bookID := chi.URLParam(r, "id")

	if err := h.service.DeleteMyReview(r.Context(), bookID, claims.UserID); err != nil {
		if err.Error() == "review not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete review")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "review deleted"})
}

// POST /reviews/{id}/like
func (h *Handler) LikeReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	reviewID := chi.URLParam(r, "id")

	if err := h.service.LikeReview(r.Context(), reviewID, claims.UserID); err != nil {
		switch err.Error() {
		case "review not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "cannot like your own review", "already liked this review":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to like review")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "review liked"})
}

// DELETE /reviews/{id}/like
func (h *Handler) UnlikeReview(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	reviewID := chi.URLParam(r, "id")

	if err := h.service.UnlikeReview(r.Context(), reviewID, claims.UserID); err != nil {
		switch err.Error() {
		case "like not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to unlike review")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "review unliked"})
}
