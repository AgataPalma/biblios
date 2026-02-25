package books

import (
	"encoding/json"
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type submitBookRequest struct {
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	CoverURL    *string      `json:"cover_url"`
	Authors     []string     `json:"authors"`
	Genres      []string     `json:"genres"`
	Edition     EditionInput `json:"edition"`
	Condition   *string      `json:"condition"`
}

func (h *Handler) SubmitBook(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool

	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req submitBookRequest
	var err error = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}
	if len(req.Authors) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "at least one author is required")
		return
	}
	if req.Edition.Format == "" {
		writeError(w, http.StatusUnprocessableEntity, "edition format is required")
		return
	}
	if req.Edition.Language == "" {
		writeError(w, http.StatusUnprocessableEntity, "edition language is required")
		return
	}

	var input SubmitBookInput = SubmitBookInput{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Authors:     req.Authors,
		Genres:      req.Genres,
		Edition:     req.Edition,
		Condition:   req.Condition,
		UserID:      claims.UserID,
		UserRole:    claims.Role,
	}

	var result SubmitBookResult
	result, err = h.service.SubmitBook(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to submit book")
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
