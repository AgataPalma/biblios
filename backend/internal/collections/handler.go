package collections

import (
	"encoding/json"
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/books"
	"github.com/AgataPalma/biblios/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, col)
}

// GET /libraries/{id}/collections
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	cols, err := h.service.ListByLibrary(r.Context(), libraryID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list collections")
		return
	}
	if cols == nil {
		cols = []Collection{}
	}
	httpx.WriteJSON(w, http.StatusOK, cols)
}

// GET /libraries/{id}/collections/{collectionId}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "collectionId")
	col, err := h.service.Get(r.Context(), collectionID)
	if err != nil || col == nil {
		httpx.WriteError(w, http.StatusNotFound, "collection not found")
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
	collectionID := chi.URLParam(r, "collectionId")

	var input UpdateCollectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	col, err := h.service.Update(r.Context(), collectionID, claims.UserID, input)
	if err != nil {
		switch err.Error() {
		case "collection not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "only the collection creator can update it", "name cannot be empty":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update collection")
		}
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
	collectionID := chi.URLParam(r, "collectionId")

	if err := h.service.Delete(r.Context(), collectionID, claims.UserID); err != nil {
		switch err.Error() {
		case "collection not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "only the collection creator can delete it":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to delete collection")
		}
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
	collectionID := chi.URLParam(r, "collectionId")

	var req struct {
		CopyID string `json:"copy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CopyID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "copy_id is required")
		return
	}

	if err := h.service.AddBook(r.Context(), collectionID, claims.UserID, req.CopyID); err != nil {
		switch err.Error() {
		case "collection not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "only the collection creator can add books to this collection":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to add book")
		}
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
	collectionID := chi.URLParam(r, "collectionId")
	copyID := chi.URLParam(r, "copyId")

	if err := h.service.RemoveBook(r.Context(), collectionID, claims.UserID, copyID); err != nil {
		switch err.Error() {
		case "collection not found", "book not found in collection":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "only the collection creator can remove books":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to remove book")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "book removed from collection"})
}

// GET /libraries/{id}/collections/{collectionId}/books
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	collectionID := chi.URLParam(r, "collectionId")
	page, limit := httpx.PaginationParams(r)

	userBooks, total, err := h.service.ListBooks(r.Context(), collectionID, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list books")
		return
	}
	if userBooks == nil {
		userBooks = []books.UserBook{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"books": userBooks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
