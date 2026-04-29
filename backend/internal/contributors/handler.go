package contributors

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

// GET /contributors?q=&page=&limit=
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := httpx.PaginationParams(r)
	query := r.URL.Query().Get("q")

	contribs, total, err := h.service.Search(r.Context(), query, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list contributors")
		return
	}
	if contribs == nil {
		contribs = []Contributor{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"contributors": contribs,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

// GET /contributors/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	detail, err := h.service.GetDetail(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get contributor")
		return
	}
	if detail == nil {
		httpx.WriteError(w, http.StatusNotFound, "contributor not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, detail)
}

// POST /contributors
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input CreateContributorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c, err := h.service.Create(r.Context(), input, claims.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, c)
}
