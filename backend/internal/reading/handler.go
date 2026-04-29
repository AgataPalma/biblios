package reading

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

// GET /reading/challenges
func (h *Handler) ListChallenges(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	challenges, err := h.service.ListChallenges(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list challenges")
		return
	}
	if challenges == nil {
		challenges = []Challenge{}
	}
	httpx.WriteJSON(w, http.StatusOK, challenges)
}

// POST /reading/challenges
func (h *Handler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input CreateChallengeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Year == 0 {
		input.Year = time.Now().Year()
	}
	c, err := h.service.CreateChallenge(r.Context(), claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}

// DELETE /reading/challenges/{id}
func (h *Handler) DeleteChallenge(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteChallenge(r.Context(), id, claims.UserID); err != nil {
		if err.Error() == "challenge not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete challenge")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "challenge deleted"})
}

// GET /reading/challenges/{id}/progress
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	progress, err := h.service.GetProgress(r.Context(), id, claims.UserID)
	if err != nil {
		if err.Error() == "challenge not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get progress")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, progress)
}

// POST /reading/sessions
func (h *Handler) LogSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input LogSessionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.service.LogSession(r.Context(), claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, session)
}

// GET /reading/sessions?copy_id=
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	page, limit := httpx.PaginationParams(r)
	copyID := r.URL.Query().Get("copy_id")

	sessions, total, err := h.service.ListSessions(r.Context(), claims.UserID, copyID, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []Session{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GET /reading/stats
func (h *Handler) GetOverallStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	stats, err := h.service.GetOverallStats(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}

// GET /reading/stats/year/{year}
func (h *Handler) GetYearStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil || year < 2000 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid year")
		return
	}
	stats, err := h.service.GetYearStats(r.Context(), claims.UserID, year)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get year stats")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}

// GET /reading/stats/month/{year}/{month}
func (h *Handler) GetMonthStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid year")
		return
	}
	month, err := strconv.Atoi(chi.URLParam(r, "month"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid month")
		return
	}
	stats, err := h.service.GetMonthStats(r.Context(), claims.UserID, year, month)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}
