package books

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AgataPalma/biblios/internal/httpx"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/go-chi/chi/v5"
)

// LookupService is satisfied by the lookup.Service adapter in main.go.
type LookupService interface {
	GetCoverURL(ctx context.Context, isbn string) (string, error)
}

type Handler struct {
	service       *Service
	lookupService LookupService
	coversDir     string
}

func NewHandler(service *Service, lookupService LookupService, coversDir string) *Handler {
	return &Handler{service: service, lookupService: lookupService, coversDir: coversDir}
}

// ─── Request types ────────────────────────────────────────────────────────────

type submitBookRequest struct {
	Title          string       `json:"title"`
	Authors        []string     `json:"authors"`
	Genres         []string     `json:"genres"`
	SeriesName     *string      `json:"series_name"`
	SeriesPosition *float64     `json:"series_position"`
	Edition        EditionInput `json:"edition"`
	Condition      *string      `json:"condition"`
	CatalogueOnly  bool         `json:"catalogue_only"`
	ReadingStatus  string       `json:"reading_status"`
	CurrentPage    *int         `json:"current_page"`
	OwnedByUser    *bool        `json:"owned_by_user"`
	BorrowedFrom   *string      `json:"borrowed_from"`
	Location       *string      `json:"location"`
}

type addCopyRequest struct {
	EditionID     string  `json:"edition_id"`
	Condition     *string `json:"condition"`
	ReadingStatus string  `json:"reading_status"`
	CurrentPage   *int    `json:"current_page"`
	OwnedByUser   *bool   `json:"owned_by_user"`
	BorrowedFrom  *string `json:"borrowed_from"`
	Location      *string `json:"location"`
}

type updateCopyStatusRequest struct {
	Status      string `json:"status"`
	CurrentPage *int   `json:"current_page"`
}

type updateBookRequest struct {
	Title          *string       `json:"title"`
	Description    *string       `json:"description"`
	Authors        []string      `json:"authors"`
	Genres         []string      `json:"genres"`
	SeriesName     *string       `json:"series_name"`
	SeriesPosition *float64      `json:"series_position"`
	EditionID      string        `json:"edition_id"`
	Edition        *EditionInput `json:"edition"`
}

// ─── POST /books ──────────────────────────────────────────────────────────────

func (h *Handler) SubmitBook(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req submitBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}
	if len(req.Authors) == 0 {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "at least one author is required")
		return
	}
	if req.Edition.Format == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "edition.format is required")
		return
	}
	if req.Edition.Language == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "edition.language is required")
		return
	}

	input := SubmitBookInput{
		Title:          req.Title,
		Authors:        req.Authors,
		Genres:         req.Genres,
		SeriesName:     req.SeriesName,
		SeriesPosition: req.SeriesPosition,
		Edition:        req.Edition,
		Condition:      req.Condition,
		UserID:         claims.UserID,
		UserRole:       claims.Role,
		CatalogueOnly:  req.CatalogueOnly,
		CopyOptions: CopyOptions{
			ReadingStatus: req.ReadingStatus,
			CurrentPage:   req.CurrentPage,
			OwnedByUser:   req.OwnedByUser,
			BorrowedFrom:  req.BorrowedFrom,
			Location:      req.Location,
		},
	}

	result, err := h.service.SubmitBook(r.Context(), input)
	if err != nil {
		slog.Error("submit book failed", "error", err, "user_id", claims.UserID)
		switch {
		case strings.Contains(err.Error(), "invalid ISBN"):
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		case err.Error() == "edition with this ISBN already exists":
			var isbn string
			if req.Edition.ISBN13 != nil {
				isbn = *req.Edition.ISBN13
			} else if req.Edition.ISBN10 != nil {
				isbn = *req.Edition.ISBN10
			} else if req.Edition.ISBN != nil {
				isbn = *req.Edition.ISBN
			}
			existing, _ := h.service.FindExistingEditionByISBN(r.Context(), isbn)
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{
				"error":   "edition already exists",
				"edition": existing,
			})
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to submit book")
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

// ─── GET /books ───────────────────────────────────────────────────────────────

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := httpx.PaginationParams(r)

	// If a search query is present, use full search with filters
	if q.Get("q") != "" || q.Get("format") != "" || q.Get("language") != "" ||
		q.Get("genre") != "" || q.Get("series") != "" || q.Get("publisher") != "" ||
		q.Get("year_from") != "" || q.Get("year_to") != "" || q.Get("award") != "" || q.Get("mood") != "" {

		filters := SearchFilters{
			Query:     q.Get("q"),
			Format:    q.Get("format"),
			Language:  q.Get("language"),
			Genre:     q.Get("genre"),
			Series:    q.Get("series"),
			Publisher: q.Get("publisher"),
			Award:     q.Get("award"),
			Mood:      q.Get("mood"),
			Sort:      q.Get("sort"),
			Page:      page,
			Limit:     limit,
		}
		if yf := q.Get("year_from"); yf != "" {
			if v, err := strconv.Atoi(yf); err == nil {
				filters.YearFrom = v
			}
		}
		if yt := q.Get("year_to"); yt != "" {
			if v, err := strconv.Atoi(yt); err == nil {
				filters.YearTo = v
			}
		}

		result, err := h.service.SearchBooks(r.Context(), filters)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "search failed")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, result)
		return
	}

	result, err := h.service.ListBooks(r.Context(), page, limit, q.Get("sort"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list books")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// ─── GET /books/check?isbn= ───────────────────────────────────────────────────

func (h *Handler) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	isbnParam := r.URL.Query().Get("isbn")
	if isbnParam == "" {
		httpx.WriteError(w, http.StatusBadRequest, "isbn query parameter is required")
		return
	}
	edition, err := h.service.FindExistingEditionByISBN(r.Context(), isbnParam)
	if err != nil {
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if edition == nil {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"exists": false})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"exists": true, "edition": edition})
}

// ─── GET /books/{id} ──────────────────────────────────────────────────────────

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	book, err := h.service.GetBook(r.Context(), id)
	if err != nil || book == nil {
		httpx.WriteError(w, http.StatusNotFound, "book not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, book)
}

// ─── PUT /books/{id} ─────────────────────────────────────────────────────────

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := UpdateBookInput{
		ID:             id,
		Title:          req.Title,
		Description:    req.Description,
		Authors:        req.Authors,
		Genres:         req.Genres,
		SeriesName:     req.SeriesName,
		SeriesPosition: req.SeriesPosition,
		EditionID:      req.EditionID,
		Edition:        req.Edition,
	}

	if err := h.service.UpdateBook(r.Context(), input); err != nil {
		switch err.Error() {
		case "title cannot be empty":
			httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		case "book not found", "edition not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update book")
		}
		return
	}

	book, err := h.service.GetBook(r.Context(), id)
	if err != nil || book == nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to reload book")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, book)
}

// ─── DELETE /books/{id} ───────────────────────────────────────────────────────

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteBook(r.Context(), id); err != nil {
		switch err.Error() {
		case "book not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "cannot delete book with existing copies":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to delete book")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "book deleted"})
}

// ─── POST /books/copies ───────────────────────────────────────────────────────

func (h *Handler) AddCopy(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req addCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.EditionID == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "edition_id is required")
		return
	}

	result, err := h.service.AddCopyOfExistingEdition(r.Context(), AddCopyInput{
		EditionID: req.EditionID,
		Condition: req.Condition,
		UserID:    claims.UserID,
		CopyOptions: CopyOptions{
			ReadingStatus: req.ReadingStatus,
			CurrentPage:   req.CurrentPage,
			OwnedByUser:   req.OwnedByUser,
			BorrowedFrom:  req.BorrowedFrom,
			Location:      req.Location,
		},
	})
	if err != nil {
		switch err.Error() {
		case "edition not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		case "edition is not yet approved":
			httpx.WriteError(w, http.StatusConflict, err.Error())
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "failed to add copy")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, result)
}

// ─── GET /users/me/books ──────────────────────────────────────────────────────

func (h *Handler) GetMyBooks(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, limit := httpx.PaginationParams(r)
	q := r.URL.Query()

	books, total, err := h.service.GetUserBooks(
		r.Context(), claims.UserID, page, limit,
		q.Get("status"), q.Get("genre"), q.Get("q"), q.Get("sort"),
	)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to get books")
		return
	}
	if books == nil {
		books = []UserBook{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"books": books,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ─── PUT /books/copies/{id}/status ───────────────────────────────────────────

func (h *Handler) UpdateCopyStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")
	var req updateCopyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Status == "" {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "status is required")
		return
	}

	if err := h.service.UpdateCopyStatus(r.Context(), copyID, claims.UserID, req.Status, req.CurrentPage); err != nil {
		switch err.Error() {
		case "copy not found":
			httpx.WriteError(w, http.StatusNotFound, err.Error())
		default:
			if strings.Contains(err.Error(), "invalid reading status") {
				httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update status")
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// ─── DELETE /books/copies/{id} ────────────────────────────────────────────────

func (h *Handler) RemoveCopy(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	copyID := chi.URLParam(r, "id")
	if err := h.service.RemoveCopy(r.Context(), copyID, claims.UserID); err != nil {
		if err.Error() == "copy not found" {
			httpx.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "failed to remove copy")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "copy removed"})
}

// ─── POST /books/{id}/cover ───────────────────────────────────────────────────

func (h *Handler) UploadCover(w http.ResponseWriter, r *http.Request) {
	editionID := chi.URLParam(r, "id")

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("cover")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cover field is required")
		return
	}
	defer file.Close()

	// Validate content type / extension
	allowed := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext, ok := allowed[header.Header.Get("Content-Type")]
	if !ok {
		ext = strings.ToLower(filepath.Ext(header.Filename))
		if ext == ".jpeg" {
			ext = ".jpg"
		}
		if ext != ".jpg" && ext != ".png" && ext != ".webp" {
			httpx.WriteError(w, http.StatusUnprocessableEntity, "only JPEG, PNG and WebP images are accepted")
			return
		}
	}

	if err = os.MkdirAll(h.coversDir, 0755); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create covers directory")
		return
	}

	filename := editionID + ext
	dst, err := os.Create(filepath.Join(h.coversDir, filename))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to save cover image")
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to write cover image")
		return
	}

	publicURL := fmt.Sprintf("/covers/%s", filename)
	if err = h.service.UpdateEditionCoverURL(r.Context(), editionID, publicURL); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update cover URL")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"cover_url": publicURL})
}

// ─── POST /admin/backfill-covers ─────────────────────────────────────────────

func (h *Handler) BackfillCovers(w http.ResponseWriter, r *http.Request) {
	books, err := h.service.GetBooksWithoutCovers(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch books")
		return
	}

	updated := 0
	for _, book := range books {
		for _, edition := range book.Editions {
			isbnStr := edition.PreferredISBN()
			if isbnStr == "" {
				continue
			}
			coverURL, err := h.lookupService.GetCoverURL(r.Context(), isbnStr)
			if err != nil || coverURL == "" {
				continue
			}
			if err = h.service.UpdateEditionCoverURL(r.Context(), edition.ID, coverURL); err == nil {
				updated++
				break
			}
		}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"total":   len(books),
		"updated": updated,
	})
}
