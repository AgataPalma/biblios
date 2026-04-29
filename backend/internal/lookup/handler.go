package lookup

import (
	"net/http"

	"github.com/AgataPalma/biblios/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Lookup handles GET /books/lookup?isbn=<isbn>
// Returns merged results from Google Books and Open Library.
func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	isbnParam := r.URL.Query().Get("isbn")
	if isbnParam == "" {
		httpx.WriteError(w, http.StatusBadRequest, "isbn query parameter is required")
		return
	}

	results, err := h.service.LookupByISBN(r.Context(), isbnParam)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if results == nil {
		results = []LookupResult{}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"results": results,
		"total":   len(results),
	})
}
