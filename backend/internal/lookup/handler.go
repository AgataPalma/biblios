package lookup

import (
	"encoding/json"
	"net/http"
	"strconv"
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

	if isbn != "" {
		var result *GoogleBooksResult
		var err error
		result, err = h.service.LookupByISBN(r.Context(), isbn)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		if result == nil {
			writeError(w, http.StatusNotFound, "book not found in external APIs")
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	if title != "" {
		var page int = 1
		var err error
		if p := r.URL.Query().Get("page"); p != "" {
			page, err = strconv.Atoi(p)
			if err != nil || page < 1 {
				page = 1
			}
		}

		var result *SearchResultList
		result, err = h.service.LookupByTitleAuthor(r.Context(), title, author, page)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		if result == nil || len(result.Results) == 0 {
			writeError(w, http.StatusNotFound, "no books found in external APIs")
			return
		}
		writeJSON(w, http.StatusOK, result)
		return
	}

	writeError(w, http.StatusBadRequest, "provide isbn or title query parameter")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
