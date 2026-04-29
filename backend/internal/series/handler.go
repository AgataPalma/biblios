package series

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

// GET /series?q=&page=&limit=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := httpx.PaginationParams(r)
	query := r.URL.Query().Get("q")

	results, total, err := h.service.Search(r.Context(), query, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list series")
		return
	}
	if results == nil {
		results = []Series{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"series": results,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GET /series/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	detail, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get series")
		return
	}
	if detail == nil {
		httpx.WriteError(w, http.StatusNotFound, "series not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, detail)
}

// POST /series
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input CreateSeriesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	s, err := h.service.Create(r.Context(), input, claims.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, s)
}
