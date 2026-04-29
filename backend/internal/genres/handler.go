package genres

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

// GET /genres
func (h *Handler) ListGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.service.ListGenres(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list genres")
		return
	}
	if genres == nil {
		genres = []Genre{}
	}
	httpx.WriteJSON(w, http.StatusOK, genres)
}

// POST /genres
func (h *Handler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	g, err := h.service.CreateGenre(r.Context(), req.Name, claims.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, g)
}

// GET /moods
func (h *Handler) ListMoods(w http.ResponseWriter, r *http.Request) {
	moods, err := h.service.ListMoods(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list moods")
		return
	}
	if moods == nil {
		moods = []Mood{}
	}
	httpx.WriteJSON(w, http.StatusOK, moods)
}

// POST /moods
func (h *Handler) CreateMood(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	m, err := h.service.CreateMood(r.Context(), req.Name, claims.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, m)
}
