package books

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"

	"github.com/AgataPalma/biblios/internal/apictx"
)

type LookupService interface {
	LookupByISBN(ctx context.Context, isbn string) (string, error)
}

type Handler struct {
	service       *Service
	lookupService LookupService
}

func NewHandler(service *Service, lookupService LookupService) *Handler {
	return &Handler{
		service:       service,
		lookupService: lookupService,
	}
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

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	var page int = 1
	var limit int = 20
	var err error

	if p := r.URL.Query().Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil || page < 1 {
			page = 1
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	var result ListBooksResult
	result, err = h.service.ListBooks(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list books")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	var id string = chi.URLParam(r, "id")

	var book *Book
	var err error
	book, err = h.service.GetBook(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}

	writeJSON(w, http.StatusOK, book)
}

type updateBookRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CoverURL    *string `json:"cover_url"`
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	var id string = chi.URLParam(r, "id")

	var req updateBookRequest
	var err error = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.UpdateBook(r.Context(), UpdateBookInput{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
	})
	if err != nil {
		if err.Error() == "title is required" {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if err.Error() == "book not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update book")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "book updated"})
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	var id string = chi.URLParam(r, "id")

	var err error = h.service.DeleteBook(r.Context(), id)
	if err != nil {
		if err.Error() == "book not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "cannot delete book with existing copies" {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "book deleted"})
}

func (h *Handler) GetMyBooks(w http.ResponseWriter, r *http.Request) {
	var claims apictx.Claims
	var ok bool
	claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var page int = 1
	var limit int = 20
	var err error

	if p := r.URL.Query().Get("page"); p != "" {
		page, err = strconv.Atoi(p)
		if err != nil || page < 1 {
			page = 1
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 || limit > 100 {
			limit = 20
		}
	}

	var result ListBooksResult
	result, err = h.service.GetUserBooks(r.Context(), claims.UserID, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user books")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) BackfillCovers(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.GetBooksWithoutCovers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}

	updated := 0
	for _, book := range books {
		for _, edition := range book.Editions {
			if edition.ISBN == nil || *edition.ISBN == "" {
				continue
			}
			coverURL, err := h.lookupService.LookupByISBN(r.Context(), *edition.ISBN)
			if err != nil || coverURL == "" {
				continue
			}
			err = h.service.UpdateCoverURL(r.Context(), book.ID, coverURL)
			if err == nil {
				updated++
				break
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":   len(books),
		"updated": updated,
	})
}

type updateReadingStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateReadingStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")

	var req updateReadingStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	valid := map[string]bool{"want_to_read": true, "reading": true, "read": true}
	if !valid[req.Status] {
		writeError(w, http.StatusUnprocessableEntity, "invalid status")
		return
	}

	if err := h.service.UpdateReadingStatus(r.Context(), copyID, claims.UserID, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) RemoveCopy(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")

	if err := h.service.RemoveCopy(r.Context(), copyID, claims.UserID); err != nil {
		if err.Error() == "copy not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to remove copy")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "copy removed"})
}

func (h *Handler) GetMyLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	books, total, err := h.service.GetMyLibrary(r.Context(), claims.UserID, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get library")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"books": books,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
