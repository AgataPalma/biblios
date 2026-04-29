package library

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

// ─── POST /libraries ──────────────────────────────────────────────────────────

func (h *Handler) CreateLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input CreateLibraryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	lib, err := h.service.CreateLibrary(r.Context(), claims.UserID, input)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, lib)
}

// ─── GET /libraries ───────────────────────────────────────────────────────────

func (h *Handler) ListMyLibraries(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libs, err := h.service.ListMyLibraries(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list libraries")
		return
	}
	if libs == nil {
		libs = []Library{}
	}
	httpx.WriteJSON(w, http.StatusOK, libs)
}

// ─── GET /libraries/public ────────────────────────────────────────────────────

func (h *Handler) ListPublicLibraries(w http.ResponseWriter, r *http.Request) {
	page, limit := httpx.PaginationParams(r)
	libs, total, err := h.service.ListPublicLibraries(r.Context(), page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list libraries")
		return
	}
	if libs == nil {
		libs = []Library{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"libraries": libs,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// ─── GET /libraries/{id} ──────────────────────────────────────────────────────

func (h *Handler) GetLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	lib, err := h.service.GetLibrary(r.Context(), id, claims.UserID)
	if err != nil {
		switch err.Error() {
		case "library not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "access denied":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to get library")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lib)
}

// ─── PUT /libraries/{id} ──────────────────────────────────────────────────────

func (h *Handler) UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	var input UpdateLibraryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	lib, err := h.service.UpdateLibrary(r.Context(), id, claims.UserID, input)
	if err != nil {
		switch err.Error() {
		case "library not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "only the library owner can update library settings", "name cannot be empty", "invalid visibility":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		case "not a member of this library":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update library")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, lib)
}

// ─── DELETE /libraries/{id} ───────────────────────────────────────────────────

func (h *Handler) DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")

	// Only owner can delete
	m, err := h.service.repo.GetMember(r.Context(), id, claims.UserID)
	if err != nil || m == nil || !m.IsOwner {
		httpx.WriteError(w, http.StatusForbidden, "only the library owner can delete this library")
		return
	}
	if err := h.service.repo.DeleteLibrary(r.Context(), id); err != nil {
		if err.Error() == "library not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete library")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "library deleted"})
}

// ─── GET /libraries/{id}/members ─────────────────────────────────────────────

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	members, err := h.service.ListMembers(r.Context(), id, claims.UserID)
	if err != nil {
		switch err.Error() {
		case "not a member of this library":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list members")
		}
		return
	}
	if members == nil {
		members = []LibraryMember{}
	}
	httpx.WriteJSON(w, http.StatusOK, members)
}

// ─── PUT /libraries/{id}/members/{userId} ─────────────────────────────────────

func (h *Handler) UpdateMemberPermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")

	var perms LibraryMember
	if err := json.NewDecoder(r.Body).Decode(&perms); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateMemberPermissions(r.Context(), libraryID, claims.UserID, targetUserID, perms); err != nil {
		switch err.Error() {
		case "not a member of this library", "you do not have permission to manage members":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		case "target user is not a member of this library", "cannot change owner permissions":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update permissions")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "permissions updated"})
}

// ─── DELETE /libraries/{id}/members/{userId} ──────────────────────────────────

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "userId")

	if err := h.service.RemoveMember(r.Context(), libraryID, claims.UserID, targetUserID); err != nil {
		switch err.Error() {
		case "not a member of this library", "only the library owner can remove members":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		case "user is not a member of this library", "cannot remove the library owner", "owner cannot remove themselves":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to remove member")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

// ─── POST /libraries/{id}/invite ─────────────────────────────────────────────

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inv, err := h.service.InviteMember(r.Context(), libraryID, claims.UserID, req.Email)
	if err != nil {
		switch err.Error() {
		case "not a member of this library", "you do not have permission to invite members":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		case "email is required":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create invitation")
		}
		return
	}
	// Strip token from response — it should be sent via email, not returned in API
	inv.Token = ""
	httpx.WriteJSON(w, http.StatusCreated, inv)
}

// ─── GET /invitations ─────────────────────────────────────────────────────────

func (h *Handler) ListMyInvitations(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	invs, err := h.service.ListMyInvitations(r.Context(), claims.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}
	if invs == nil {
		invs = []LibraryInvitation{}
	}
	httpx.WriteJSON(w, http.StatusOK, invs)
}

// ─── POST /invitations/{token}/accept ────────────────────────────────────────

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	token := chi.URLParam(r, "token")
	if err := h.service.AcceptInvitation(r.Context(), token, claims.UserID); err != nil {
		switch err.Error() {
		case "invitation not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "invitation is no longer pending", "invitation has expired":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to accept invitation")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "invitation accepted"})
}

// ─── POST /invitations/{token}/decline ───────────────────────────────────────

func (h *Handler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if err := h.service.DeclineInvitation(r.Context(), token); err != nil {
		switch err.Error() {
		case "invitation not found or already actioned":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to decline invitation")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "invitation declined"})
}

// ─── GET /libraries/{id}/books ────────────────────────────────────────────────

func (h *Handler) ListLibraryBooks(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	page, limit := httpx.PaginationParams(r)

	userBooks, total, err := h.service.ListLibraryBooks(r.Context(), libraryID, claims.UserID, page, limit)
	if err != nil {
		switch err.Error() {
		case "library not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "access denied":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list books")
		}
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

// ─── POST /libraries/{id}/books ───────────────────────────────────────────────

func (h *Handler) AddBookToLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")

	var req struct {
		CopyID string `json:"copy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CopyID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "copy_id is required")
		return
	}

	if err := h.service.AddBookToLibrary(r.Context(), libraryID, claims.UserID, req.CopyID); err != nil {
		switch err.Error() {
		case "not a member of this library", "you do not have permission to add books to this library":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to add book")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"message": "book added to library"})
}

// ─── DELETE /libraries/{id}/books/{copyId} ────────────────────────────────────

func (h *Handler) RemoveBookFromLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	libraryID := chi.URLParam(r, "id")
	copyID := chi.URLParam(r, "copyId")

	if err := h.service.RemoveBookFromLibrary(r.Context(), libraryID, claims.UserID, copyID); err != nil {
		switch err.Error() {
		case "not a member of this library", "you do not have permission to remove books from this library":
			httpx.WriteError(w, http.StatusForbidden, err.Error())
		case "book copy not found in library":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to remove book")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "book removed from library"})
}

// ─── GET /users/me/library ────────────────────────────────────────────────────
// Returns the user's first (default) library with its books.

func (h *Handler) GetMyLibrary(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	libs, err := h.service.ListMyLibraries(r.Context(), claims.UserID)
	if err != nil || len(libs) == 0 {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"library": nil, "books": []any{}, "total": 0})
		return
	}

	// Default library is the first one (created at registration)
	lib := libs[0]
	page, limit := httpx.PaginationParams(r)
	userBooks, total, err := h.service.ListLibraryBooks(r.Context(), lib.ID, claims.UserID, page, limit)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get library books")
		return
	}
	if userBooks == nil {
		userBooks = []books.UserBook{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"library": lib,
		"books":   userBooks,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
