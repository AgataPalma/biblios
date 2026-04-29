package shelves

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

// GET /shelves
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	shelves, err := h.service.ListMyShelf(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list shelves")
		return
	}
	if shelves == nil {
		shelves = []Shelf{}
	}
	httpx.WriteJSON(w, http.StatusOK, shelves)
}

// POST /shelves
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	shelf, err := h.service.Create(r.Context(), claims.UserID, req.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, shelf)
}

// PUT /shelves/{id}
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	shelf, err := h.service.Rename(r.Context(), id, claims.UserID, req.Name)
	if err != nil {
		switch err.Error() {
		case "shelf not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "name cannot be empty":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to rename shelf")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, shelf)
}

// DELETE /shelves/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), id, claims.UserID); err != nil {
		if err.Error() == "shelf not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete shelf")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "shelf deleted"})
}

// GET /shelves/{id}/books
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	page, limit := httpx.PaginationParams(r)

	userBooks, total, err := h.service.ListBooks(r.Context(), id, claims.UserID, page, limit)
	if err != nil {
		if err.Error() == "shelf not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list books")
		return
	}
	if userBooks == nil {
		userBooks = []interface{}{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"books": userBooks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// POST /shelves/{id}/books
func (h *Handler) AddBook(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var req struct {
		CopyID string `json:"copy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CopyID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "copy_id is required")
		return
	}
	if err := h.service.AddBook(r.Context(), id, claims.UserID, req.CopyID); err != nil {
		if err.Error() == "shelf not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to add book")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"message": "book added to shelf"})
}

// DELETE /shelves/{id}/books/{copyId}
func (h *Handler) RemoveBook(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	copyID := chi.URLParam(r, "copyId")

	if err := h.service.RemoveBook(r.Context(), id, claims.UserID, copyID); err != nil {
		switch err.Error() {
		case "shelf not found", "book not found on shelf":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to remove book")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "book removed from shelf"})
}
