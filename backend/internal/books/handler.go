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

func (h *Handler) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	var isbn string = r.URL.Query().Get("isbn")
	if isbn == "" {
		writeError(w, http.StatusBadRequest, "isbn query parameter is required")
		return
	}

	var edition *Edition
	var err error
	edition, err = h.service.FindExistingEditionByISBN(r.Context(), isbn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check for duplicates")
		return
	}

	if edition == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"exists": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"exists":  true,
		"edition": edition,
	})
}

type addCopyRequest struct {
	EditionID string  `json:"edition_id"`
	Condition *string `json:"condition"`
}

func (h *Handler) AddCopy(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool

	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req addCopyRequest
	var err error = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EditionID == "" {
		writeError(w, http.StatusUnprocessableEntity, "edition_id is required")
		return
	}

	var input AddCopyInput = AddCopyInput{
		EditionID: req.EditionID,
		Condition: req.Condition,
		UserID:    claims.UserID,
	}

	var result AddCopyResult
	result, err = h.service.AddCopyOfExistingEdition(r.Context(), input)
	if err != nil {
		if err.Error() == "edition is not yet approved" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if err.Error() == "edition not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to add copy")
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
