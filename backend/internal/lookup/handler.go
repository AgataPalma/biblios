package lookup

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Lookup(w http.ResponseWriter, r *http.Request) {
	var isbn string = r.URL.Query().Get("isbn")
	var title string = r.URL.Query().Get("title")
	var author string = r.URL.Query().Get("author")

	var result *GoogleBooksResult
	var err error

	if isbn != "" {
		result, err = h.service.LookupByISBN(r.Context(), isbn)
	} else if title != "" {
		result, err = h.service.LookupByTitleAuthor(r.Context(), title, author)
	} else {
		writeError(w, http.StatusBadRequest, "provide isbn or title query parameter")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup failed")
		return
	}

	if result == nil {
		writeError(w, http.StatusNotFound, "book not found in external APIs")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
