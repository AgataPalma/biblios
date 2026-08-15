package collections

import (
	"encoding/json"
	"errors"
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

func writeCollectionServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, ErrLibraryNotFound), errors.Is(err, ErrCollectionNotFound):
		httpx.WriteError(w, http.StatusNotFound, "collection not found")
	case errors.Is(err, ErrAccessDenied):
		httpx.WriteError(w, http.StatusForbidden, "access denied")
	case errors.Is(err, ErrCopyNotInLibrary), errors.Is(err, ErrNameRequired), errors.Is(err, ErrNameEmpty):
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrBookNotInCollection):
		httpx.WriteError(w, http.StatusNotFound, err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, fallback)
	}
}

// POST /libraries/{id}/collections
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")

	var input CreateCollectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	col, err := h.service.Create(r.Context(), libraryID, claims.UserID, input)
	if err != nil {
		writeCollectionServiceError(w, err, "failed to create collection")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, col)
}

// GET /libraries/{id}/collections
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	cols, err := h.service.ListByLibrary(r.Context(), libraryID, claims.UserID)
	if err != nil {
		writeCollectionServiceError(w, err, "failed to list collections")
		return
	}
	if cols == nil {
		cols = []Collection{}
	}
	httpx.WriteJSON(w, http.StatusOK, cols)
}

// GET /libraries/{id}/collections/{collectionId}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")
	col, err := h.service.Get(r.Context(), libraryID, collectionID, claims.UserID)
	if err != nil {
		writeCollectionServiceError(w, err, "failed to get collection")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, col)
}

// PUT /libraries/{id}/collections/{collectionId}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")

	var input UpdateCollectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	col, err := h.service.Update(r.Context(), libraryID, collectionID, claims.UserID, input)
	if err != nil {
		writeCollectionServiceError(w, err, "failed to update collection")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, col)
}

// DELETE /libraries/{id}/collections/{collectionId}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")

	if err := h.service.Delete(r.Context(), libraryID, collectionID, claims.UserID); err != nil {
		writeCollectionServiceError(w, err, "failed to delete collection")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "collection deleted"})
}

// POST /libraries/{id}/collections/{collectionId}/books
func (h *Handler) AddBook(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")

	var req struct {
		CopyID string `json:"copy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CopyID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "copy_id is required")
		return
	}

	if err := h.service.AddBook(r.Context(), libraryID, collectionID, claims.UserID, req.CopyID); err != nil {
		writeCollectionServiceError(w, err, "failed to add book")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"message": "book added to collection"})
}

// DELETE /libraries/{id}/collections/{collectionId}/books/{copyId}
func (h *Handler) RemoveBook(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")
	copyID := chi.URLParam(r, "copyId")

	if err := h.service.RemoveBook(r.Context(), libraryID, collectionID, claims.UserID, copyID); err != nil {
		writeCollectionServiceError(w, err, "failed to remove book")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "book removed from collection"})
}

// GET /libraries/{id}/collections/{collectionId}/books
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	collectionID := chi.URLParam(r, "collectionId")
	page, limit := httpx.PaginationParams(r)

	collectionBooks, total, err := h.service.ListBooks(r.Context(), libraryID, collectionID, claims.UserID, page, limit)
	if err != nil {
		writeCollectionServiceError(w, err, "failed to list books")
		return
	}
	if collectionBooks == nil {
		collectionBooks = []CollectionBook{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"books": collectionBooks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
